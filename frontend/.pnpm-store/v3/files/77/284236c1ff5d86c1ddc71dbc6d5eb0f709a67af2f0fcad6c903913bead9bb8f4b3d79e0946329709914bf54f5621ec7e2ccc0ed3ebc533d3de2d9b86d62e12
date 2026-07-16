type Pipe<M = any, N = any> = (msg: M, resolve?: Resolve<N>) => void;
type Resolve<M> = (msg: M) => void;

declare enum SeverityType {
    VERBOSE = "verbose",
    ERROR = "error",
    WARN = "warn",
    INFO = "info",
    HTTP = "http",
    AUTO_DETECT_ERROR = "autoDetectError",
    PERFORMANCE = "performance",
    SPEED = "speed",
    AUTO_DETECT_EVENT = "autoDetectEvent"
}
declare enum Environment {
    production = "prod",
    demo = "demo",
    staging = "staging"
}
declare enum NetworkType {
    unknown = "unknown",
    wifi = "wifi",
    net2g = "net2g",
    net3g = "net3g",
    net4g = "net4g",
    net5g = "net5g"
}
declare enum AutoDetectedEventType {
    onPageChange = "onPageChange",
    onPageRetention = "onPageRetention",
    httpClientMetrics = "httpClientMetrics"
}
interface ExtraCommonDataFormat {
    [key: string]: any;
}

type ExtraInfo = {
    [key: string]: any;
};
interface CommonData {
    appName: string;
    deviceId: string;
    sessionId: string;
    env: Environment;
    accountId?: string;
    appVersion?: string;
    networkType?: NetworkType;
    [key: string]: any;
}
type DisallowedStrings = AutoDetectedEventType.onPageChange | AutoDetectedEventType.onPageRetention;
type AllowedString<S extends string> = S extends DisallowedStrings ? never : S;

interface Config {
    appName: string;
    appVersion: string;
    env: Environment;
    delay?: number;
    deviceId?: string;
    accountId?: string;
    isWebappContainer?: boolean;
    errorRepeatTime?: number;
    enableErrorMonitoring?: boolean;
    enableDetectPageChange?: boolean;
    enableRecordPageRetention?: boolean;
    enableRecordAPIPerformance?: boolean;
    assetSpeedMonitoringWhiteList?: (string | RegExp)[];
    enableRecordAPIPerformanceWhitelist?: (string | RegExp)[];
    enableRecordAPIPerformancePercentage?: number;
    hideSearchParams?: boolean;
    useSessionStorageForSessionId?: boolean;
    getMFEName?: () => string;
    assetSpeedMonitoringWhiteListByMFE?: {
        [key: string]: (string | RegExp)[];
    };
    disableThrottle?: boolean;
}

interface AirTrackerPluginOption {
    name: string;
    setUp?: (airTracker: AirTracker) => void;
    destroy?: () => void;
    [key: string]: any;
}
declare class AirTrackerPlugin {
    name: string;
    option: AirTrackerPluginOption;
    private isInit;
    constructor(pluginOption: AirTrackerPluginOption);
    patch(airTracker: AirTracker): void;
    uninstall(): void;
}

interface InterfaceEventEmitter {
    indexOf: Function;
    on: Function;
    one: Function;
    emit: Function;
    remove: Function;
    clear: Function;
}

declare class AirTracker {
    config: Config;
    lifeCycle: InterfaceEventEmitter;
    private plugins;
    private _commonData;
    private _MFECommonDataMap;
    private timeMap;
    constructor({ config, plugins }: {
        config: Config;
        plugins?: AirTrackerPlugin[];
    });
    private installPlugins;
    get commonData(): CommonData;
    get MFECommonDataMap(): {
        [key: string]: ExtraCommonDataFormat;
    };
    private initCommonData;
    private updateCommonDataBasedOnConfig;
    updateCommonData(extraData: ExtraCommonDataFormat): void;
    removeCommonDataByKey(keys: string[]): void;
    addToMFECommonData({ MFEName, MFECommonData }: {
        MFEName: string;
        MFECommonData: ExtraCommonDataFormat;
    }): void;
    setConfig(config: Partial<Config>): void;
    private packedDataConvert;
    private normalPipelineObj;
    private observerPipelineObj;
    immediateSendPipe: Pipe<any, any>;
    observerPipeline: ({ severity, eventName, extraInfo, }: {
        severity: string;
        eventName: string;
        extraInfo?: ExtraInfo | undefined;
    }) => void;
    normalLogPipeLine: ({ severity, eventName, extraInfo, }: {
        severity: string;
        eventName: string;
        extraInfo?: ExtraInfo | undefined;
    }) => void;
    ImmediatePipeline: ({ severity, eventName, extraInfo, }: {
        severity: string;
        eventName: string;
        extraInfo?: ExtraInfo | undefined;
    }) => void;
    verbose<T extends string>(eventName: AllowedString<T>, extraInfo?: ExtraInfo): void;
    httpLog<T extends string>(eventName: AllowedString<T>, extraInfo?: ExtraInfo): void;
    info<T extends string>(eventName: AllowedString<T>, extraInfo?: ExtraInfo): void;
    warn<T extends string>(eventName: AllowedString<T>, extraInfo?: ExtraInfo): void;
    error<T extends string>(eventName: AllowedString<T>, extraInfo?: ExtraInfo): void;
    businessLog<T extends string>(eventName: AllowedString<T>, extraInfo?: ExtraInfo): void;
    reportDuration(eventName: string, duration: number, extraInfo?: ExtraInfo): void;
    private getTimerKey;
    timeStart(eventName: string): void;
    timeEnd(eventName: string, delta?: number): void;
    addToAssetSpeedWhiteListByMFE({ MFEName, whiteList }: {
        MFEName: string;
        whiteList: (string | RegExp)[];
    }): void;
    destroy(): void;
}

export { AirTrackerPlugin, type AirTrackerPluginOption, AutoDetectedEventType, type Config, Environment, NetworkType, SeverityType, AirTracker as default };
