package service

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
)

// AWS EventStream 二进制帧解码器
//
// 帧结构（big-endian）：
//
//	[total_length: 4][headers_length: 4][prelude_crc: 4][headers][payload][message_crc: 4]
//
// 同时被 Bedrock InvokeModelWithResponseStream 和 Kiro generateAssistantResponse /
// SendMessageStreaming 使用。事件类型语义由调用方根据 :event-type 头判断。

// AWSEventStreamFrame 表示一个解析后的 event stream 帧。
type AWSEventStreamFrame struct {
	EventType     string // :event-type 头（如 chunk / assistantResponseEvent）
	MessageType   string // :message-type 头（event / exception / error）
	ExceptionType string // :exception-type 头（仅在 message-type=exception 时有值）
	ContentType   string // :content-type 头（如 application/json）
	Payload       []byte // 帧的 payload 部分
}

// AWSEventStreamDecoder 流式解码 AWS event stream 二进制帧。
type AWSEventStreamDecoder struct {
	reader *bufio.Reader
}

// NewAWSEventStreamDecoder 创建一个 decoder；64KB buffered reader 与原 bedrock 实现一致。
func NewAWSEventStreamDecoder(r io.Reader) *AWSEventStreamDecoder {
	return &AWSEventStreamDecoder{reader: bufio.NewReaderSize(r, 64*1024)}
}

// NextFrame 读取下一个完整帧。EOF 时返回 io.EOF。
//
// 帧的 :message-type 为 "exception" 或 "error" 时仍然返回帧（payload 是错误体），
// 由调用方决定如何处理；CRC 失败和帧长非法直接返回 error。
func (d *AWSEventStreamDecoder) NextFrame() (*AWSEventStreamFrame, error) {
	prelude := make([]byte, 12)
	if _, err := io.ReadFull(d.reader, prelude); err != nil {
		return nil, err
	}

	preludeCRC := awsEventStreamReadUint32(prelude[8:12])
	if crc32.Checksum(prelude[0:8], awsEventStreamCRC32Table) != preludeCRC {
		return nil, fmt.Errorf("eventstream prelude CRC mismatch")
	}

	totalLength := awsEventStreamReadUint32(prelude[0:4])
	headersLength := awsEventStreamReadUint32(prelude[4:8])
	if totalLength < 16 { // 12 prelude + 4 message_crc
		return nil, fmt.Errorf("invalid eventstream frame: total_length=%d", totalLength)
	}

	remaining := int(totalLength) - 12
	data := make([]byte, remaining)
	if _, err := io.ReadFull(d.reader, data); err != nil {
		return nil, err
	}

	messageCRC := awsEventStreamReadUint32(data[len(data)-4:])
	h := crc32.New(awsEventStreamCRC32Table)
	_, _ = h.Write(prelude)
	_, _ = h.Write(data[:len(data)-4])
	if h.Sum32() != messageCRC {
		return nil, fmt.Errorf("eventstream message CRC mismatch")
	}

	if int(headersLength) > len(data)-4 {
		return nil, fmt.Errorf("invalid eventstream frame: headers_length=%d", headersLength)
	}
	headers := data[:headersLength]
	payload := data[headersLength : len(data)-4]

	return &AWSEventStreamFrame{
		EventType:     extractAWSEventStreamHeader(headers, ":event-type"),
		MessageType:   extractAWSEventStreamHeader(headers, ":message-type"),
		ExceptionType: extractAWSEventStreamHeader(headers, ":exception-type"),
		ContentType:   extractAWSEventStreamHeader(headers, ":content-type"),
		Payload:       payload,
	}, nil
}

// extractAWSEventStreamHeader 从二进制 headers 中提取指定 header 的字符串值。
//
// header 格式：[name_length: 1][name: variable][value_type: 1][value: variable]
// value_type=7 (string)：[length: 2][bytes: length]
//
// 完整支持 9 种 value_type 的字节跳转逻辑，确保不会因为遇到未识别的非字符串
// 头而错过目标 header。
func extractAWSEventStreamHeader(headers []byte, targetName string) string {
	pos := 0
	for pos < len(headers) {
		nameLen := int(headers[pos])
		pos++
		if pos+nameLen > len(headers) {
			break
		}
		name := string(headers[pos : pos+nameLen])
		pos += nameLen

		if pos >= len(headers) {
			break
		}
		valueType := headers[pos]
		pos++

		switch valueType {
		case 7: // string
			if pos+2 > len(headers) {
				return ""
			}
			valueLen := int(awsEventStreamReadUint16(headers[pos : pos+2]))
			pos += 2
			if pos+valueLen > len(headers) {
				return ""
			}
			value := string(headers[pos : pos+valueLen])
			pos += valueLen
			if name == targetName {
				return value
			}
		case 0: // bool true
			if name == targetName {
				return "true"
			}
		case 1: // bool false
			if name == targetName {
				return "false"
			}
		case 2: // byte
			pos++
		case 3: // short
			pos += 2
		case 4: // int
			pos += 4
		case 5: // long
			pos += 8
		case 6: // bytes
			if pos+2 > len(headers) {
				return ""
			}
			valueLen := int(awsEventStreamReadUint16(headers[pos : pos+2]))
			pos += 2 + valueLen
		case 8: // timestamp
			pos += 8
		case 9: // uuid
			pos += 16
		default:
			return ""
		}
	}
	return ""
}

var awsEventStreamCRC32Table = crc32.MakeTable(crc32.IEEE)

func awsEventStreamReadUint32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func awsEventStreamReadUint16(b []byte) uint16 {
	return uint16(b[0])<<8 | uint16(b[1])
}
