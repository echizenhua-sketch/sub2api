"use strict";
function _array_like_to_array(arr, len) {
    if (len == null || len > arr.length) len = arr.length;
    for(var i = 0, arr2 = new Array(len); i < len; i++)arr2[i] = arr[i];
    return arr2;
}
function _array_with_holes(arr) {
    if (Array.isArray(arr)) return arr;
}
function _array_without_holes(arr) {
    if (Array.isArray(arr)) return _array_like_to_array(arr);
}
function asyncGeneratorStep(gen, resolve, reject, _next, _throw, key, arg) {
    try {
        var info = gen[key](arg);
        var value = info.value;
    } catch (error) {
        reject(error);
        return;
    }
    if (info.done) {
        resolve(value);
    } else {
        Promise.resolve(value).then(_next, _throw);
    }
}
function _async_to_generator(fn) {
    return function() {
        var self = this, args = arguments;
        return new Promise(function(resolve, reject) {
            var gen = fn.apply(self, args);
            function _next(value) {
                asyncGeneratorStep(gen, resolve, reject, _next, _throw, "next", value);
            }
            function _throw(err) {
                asyncGeneratorStep(gen, resolve, reject, _next, _throw, "throw", err);
            }
            _next(undefined);
        });
    };
}
function _class_call_check(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
        throw new TypeError("Cannot call a class as a function");
    }
}
function _defineProperties(target, props) {
    for(var i = 0; i < props.length; i++){
        var descriptor = props[i];
        descriptor.enumerable = descriptor.enumerable || false;
        descriptor.configurable = true;
        if ("value" in descriptor) descriptor.writable = true;
        Object.defineProperty(target, descriptor.key, descriptor);
    }
}
function _create_class(Constructor, protoProps, staticProps) {
    if (protoProps) _defineProperties(Constructor.prototype, protoProps);
    if (staticProps) _defineProperties(Constructor, staticProps);
    return Constructor;
}
function _define_property(obj, key, value) {
    if (key in obj) {
        Object.defineProperty(obj, key, {
            value: value,
            enumerable: true,
            configurable: true,
            writable: true
        });
    } else {
        obj[key] = value;
    }
    return obj;
}
function _instanceof(left, right) {
    if (right != null && typeof Symbol !== "undefined" && right[Symbol.hasInstance]) {
        return !!right[Symbol.hasInstance](left);
    } else {
        return left instanceof right;
    }
}
function _iterable_to_array(iter) {
    if (typeof Symbol !== "undefined" && iter[Symbol.iterator] != null || iter["@@iterator"] != null) return Array.from(iter);
}
function _iterable_to_array_limit(arr, i) {
    var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"];
    if (_i == null) return;
    var _arr = [];
    var _n = true;
    var _d = false;
    var _s, _e;
    try {
        for(_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true){
            _arr.push(_s.value);
            if (i && _arr.length === i) break;
        }
    } catch (err) {
        _d = true;
        _e = err;
    } finally{
        try {
            if (!_n && _i["return"] != null) _i["return"]();
        } finally{
            if (_d) throw _e;
        }
    }
    return _arr;
}
function _non_iterable_rest() {
    throw new TypeError("Invalid attempt to destructure non-iterable instance.\\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function _non_iterable_spread() {
    throw new TypeError("Invalid attempt to spread non-iterable instance.\\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function _object_spread(target) {
    for(var i = 1; i < arguments.length; i++){
        var source = arguments[i] != null ? arguments[i] : {};
        var ownKeys = Object.keys(source);
        if (typeof Object.getOwnPropertySymbols === "function") {
            ownKeys = ownKeys.concat(Object.getOwnPropertySymbols(source).filter(function(sym) {
                return Object.getOwnPropertyDescriptor(source, sym).enumerable;
            }));
        }
        ownKeys.forEach(function(key) {
            _define_property(target, key, source[key]);
        });
    }
    return target;
}
function ownKeys(object, enumerableOnly) {
    var keys = Object.keys(object);
    if (Object.getOwnPropertySymbols) {
        var symbols = Object.getOwnPropertySymbols(object);
        if (enumerableOnly) {
            symbols = symbols.filter(function(sym) {
                return Object.getOwnPropertyDescriptor(object, sym).enumerable;
            });
        }
        keys.push.apply(keys, symbols);
    }
    return keys;
}
function _object_spread_props(target, source) {
    source = source != null ? source : {};
    if (Object.getOwnPropertyDescriptors) {
        Object.defineProperties(target, Object.getOwnPropertyDescriptors(source));
    } else {
        ownKeys(Object(source)).forEach(function(key) {
            Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key));
        });
    }
    return target;
}
function _sliced_to_array(arr, i) {
    return _array_with_holes(arr) || _iterable_to_array_limit(arr, i) || _unsupported_iterable_to_array(arr, i) || _non_iterable_rest();
}
function _to_consumable_array(arr) {
    return _array_without_holes(arr) || _iterable_to_array(arr) || _unsupported_iterable_to_array(arr) || _non_iterable_spread();
}
function _type_of(obj) {
    "@swc/helpers - typeof";
    return obj && typeof Symbol !== "undefined" && obj.constructor === Symbol ? "symbol" : typeof obj;
}
function _unsupported_iterable_to_array(o, minLen) {
    if (!o) return;
    if (typeof o === "string") return _array_like_to_array(o, minLen);
    var n = Object.prototype.toString.call(o).slice(8, -1);
    if (n === "Object" && o.constructor) n = o.constructor.name;
    if (n === "Map" || n === "Set") return Array.from(n);
    if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _array_like_to_array(o, minLen);
}
function _ts_generator(thisArg, body) {
    var f, y, t, g, _ = {
        label: 0,
        sent: function() {
            if (t[0] & 1) throw t[1];
            return t[1];
        },
        trys: [],
        ops: []
    };
    return g = {
        next: verb(0),
        "throw": verb(1),
        "return": verb(2)
    }, typeof Symbol === "function" && (g[Symbol.iterator] = function() {
        return this;
    }), g;
    function verb(n) {
        return function(v) {
            return step([
                n,
                v
            ]);
        };
    }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while(_)try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [
                op[0] & 2,
                t.value
            ];
            switch(op[0]){
                case 0:
                case 1:
                    t = op;
                    break;
                case 4:
                    _.label++;
                    return {
                        value: op[1],
                        done: false
                    };
                case 5:
                    _.label++;
                    y = op[1];
                    op = [
                        0
                    ];
                    continue;
                case 7:
                    op = _.ops.pop();
                    _.trys.pop();
                    continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) {
                        _ = 0;
                        continue;
                    }
                    if (op[0] === 3 && (!t || op[1] > t[0] && op[1] < t[3])) {
                        _.label = op[1];
                        break;
                    }
                    if (op[0] === 6 && _.label < t[1]) {
                        _.label = t[1];
                        t = op;
                        break;
                    }
                    if (t && _.label < t[2]) {
                        _.label = t[2];
                        _.ops.push(op);
                        break;
                    }
                    if (t[2]) _.ops.pop();
                    _.trys.pop();
                    continue;
            }
            op = body.call(thisArg, _);
        } catch (e) {
            op = [
                6,
                e
            ];
            y = 0;
        } finally{
            f = t = 0;
        }
        if (op[0] & 5) throw op[1];
        return {
            value: op[0] ? op[1] : void 0,
            done: true
        };
    }
}
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = function(target, all) {
    for(var name in all)__defProp(target, name, {
        get: all[name],
        enumerable: true
    });
};
var __copyProps = function(to, from, except, desc) {
    if (from && (typeof from === "undefined" ? "undefined" : _type_of(from)) === "object" || typeof from === "function") {
        var _iteratorNormalCompletion = true, _didIteratorError = false, _iteratorError = undefined;
        try {
            var _loop = function() {
                var key = _step.value;
                if (!__hasOwnProp.call(to, key) && key !== except) __defProp(to, key, {
                    get: function() {
                        return from[key];
                    },
                    enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable
                });
            };
            for(var _iterator = __getOwnPropNames(from)[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true)_loop();
        } catch (err) {
            _didIteratorError = true;
            _iteratorError = err;
        } finally{
            try {
                if (!_iteratorNormalCompletion && _iterator.return != null) {
                    _iterator.return();
                }
            } finally{
                if (_didIteratorError) {
                    throw _iteratorError;
                }
            }
        }
    }
    return to;
};
var __toCommonJS = function(mod) {
    return __copyProps(__defProp({}, "__esModule", {
        value: true
    }), mod);
};
// src/index.ts
var src_exports = {};
__export(src_exports, {
    AirTrackerPlugin: function() {
        return AirTrackerPlugin;
    },
    AutoDetectedEventType: function() {
        return AutoDetectedEventType;
    },
    Environment: function() {
        return Environment;
    },
    NetworkType: function() {
        return NetworkType;
    },
    SeverityType: function() {
        return SeverityType;
    },
    default: function() {
        return src_default;
    }
});
module.exports = __toCommonJS(src_exports);
// src/constant/common.ts
var SeverityType = /* @__PURE__ */ function(SeverityType2) {
    SeverityType2["VERBOSE"] = "verbose";
    SeverityType2["ERROR"] = "error";
    SeverityType2["WARN"] = "warn";
    SeverityType2["INFO"] = "info";
    SeverityType2["HTTP"] = "http";
    SeverityType2["AUTO_DETECT_ERROR"] = "autoDetectError";
    SeverityType2["PERFORMANCE"] = "performance";
    SeverityType2["SPEED"] = "speed";
    SeverityType2["AUTO_DETECT_EVENT"] = "autoDetectEvent";
    return SeverityType2;
}(SeverityType || {});
var Environment = /* @__PURE__ */ function(Environment2) {
    Environment2["production"] = "prod";
    Environment2["demo"] = "demo";
    Environment2["staging"] = "staging";
    return Environment2;
}(Environment || {});
var NetworkType = /* @__PURE__ */ function(NetworkType2) {
    NetworkType2["unknown"] = "unknown";
    NetworkType2["wifi"] = "wifi";
    NetworkType2["net2g"] = "net2g";
    NetworkType2["net3g"] = "net3g";
    NetworkType2["net4g"] = "net4g";
    NetworkType2["net5g"] = "net5g";
    return NetworkType2;
}(NetworkType || {});
var AutoDetectedEventType = /* @__PURE__ */ function(AutoDetectedEventType2) {
    AutoDetectedEventType2["onPageChange"] = "onPageChange";
    AutoDetectedEventType2["onPageRetention"] = "onPageRetention";
    AutoDetectedEventType2["httpClientMetrics"] = "httpClientMetrics";
    return AutoDetectedEventType2;
}(AutoDetectedEventType || {});
// src/constant/config.ts
var UNKNOWN_VALUE = "airTracker_unknown";
var DEFAULT_CONFIG = {
    appName: UNKNOWN_VALUE,
    appVersion: UNKNOWN_VALUE,
    env: "staging" /* staging */ ,
    isWebappContainer: false,
    delay: 2e3,
    errorRepeatTime: 3,
    enableErrorMonitoring: false,
    assetSpeedMonitoringWhiteList: [],
    enableDetectPageChange: false,
    enableRecordPageRetention: false,
    hideSearchParams: false,
    useSessionStorageForSessionId: false,
    assetSpeedMonitoringWhiteListByMFE: {}
};
var DEFAULT_COMMON_DATA = {
    appName: UNKNOWN_VALUE,
    env: "staging" /* staging */ ,
    sessionId: UNKNOWN_VALUE,
    deviceId: UNKNOWN_VALUE
};
// src/util/jsonHelper.ts
var isJSON = function(str) {
    if (!str) {
        return false;
    }
    if (/^\s*$/.test(str)) return false;
    str = str.replace(/\\(?:["\\\/bfnrt]|u[0-9a-fA-F]{4})/g, "@");
    str = str.replace(/"[^"\\\n\r]*"|true|false|null|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?/g, "]");
    str = str.replace(/(?:^|:|,)(?:\s*\[)+/g, "");
    return /^[\],:{}\s]*$/.test(str);
};
var safeReplacer = function() {
    var seen = /* @__PURE__ */ new WeakSet();
    return function(key, value) {
        if (_instanceof(value, Error)) {
            return "Error.message: ".concat(value.message, " \n  Error.stack: ").concat(value.stack);
        }
        if ((typeof value === "undefined" ? "undefined" : _type_of(value)) === "object" && value !== null) {
            if (seen.has(value)) {
                return "[Circular ".concat(key || "root", "]");
            }
            seen.add(value);
        }
        if (typeof value === "function") {
            return "function";
        }
        if ((typeof value === "undefined" ? "undefined" : _type_of(value)) === "symbol") {
            return "symbol";
        }
        if (typeof value === "undefined") {
            return null;
        }
        return value;
    };
};
var safeStringify = function(obj) {
    if (typeof obj === "string") {
        return obj;
    }
    try {
        if (_instanceof(obj, Error)) {
            return (JSON.stringify(obj, safeReplacer()) || "undefined").replace(/"/gim, "");
        }
        return JSON.stringify(obj, safeReplacer());
    } catch (e) {
        return '{"error":"error happen when airTracker stringify"}';
    }
};
// src/util/http.ts
var request = function(param) {
    var _param_method = param.method, method = _param_method === void 0 ? "post" : _param_method, url = param.url, data = param.data, success = param.success, fail = param.fail;
    if (!isJSON(data)) {
        return;
    }
    var xhr = new XMLHttpRequest();
    xhr.addEventListener("readystatechange", function() {
        if (xhr.readyState === 4) {
            if (xhr.status >= 400 || xhr.status === 0) {
                fail === null || fail === void 0 ? void 0 : fail(xhr.response);
            } else {
                success === null || success === void 0 ? void 0 : success(xhr.response);
            }
        }
    });
    xhr.open(method, url);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.send(data);
};
// src/constant/lifeCycleEventName.ts
var lifeCycleEventName = {
    onInit: "onInit",
    onConfigInit: "onConfigInit",
    onConfigUpdated: "onConfigUpdated",
    onCommonDataInit: "onCommonDataInit",
    onCommonUpdated: "onCommonUpdated",
    onPageChange: "onPageChange",
    onDestroy: "onDestroy,"
};
// src/constant/index.ts
var DEVICE_ID_STORAGE_KEY = "AIR_ANALYTICS_DEVICE_ID";
var SESSION_ID_STORAGE_KEY = "AIR_ANALYTICS_SESSION_ID";
var RISK_DEVICE_ID_STORAGE_KEY = "AWX_RISK_ID";
var loggingServiceUrl = function(env) {
    if (env == "prod" /* production */ ) {
        return "https://o11y.airwallex.com/airtracker/logs";
    }
    if (env == "demo" /* demo */ ) {
        return "https://o11y-demo.airwallex.com/airtracker/logs";
    }
    return "https://o11y-staging.airwallex.com/airtracker/logs";
};
var loggingServiceMetricsUrl = function(env) {
    if (env == "prod" /* production */ ) {
        return "https://o11y.airwallex.com/airtracker/metrics";
    }
    if (env == "demo" /* demo */ ) {
        return "https://o11y-demo.airwallex.com/airtracker/metrics";
    }
    return "https://o11y-staging.airwallex.com/airtracker/metrics";
};
// src/util/index.ts
var generateUId = function() {
    var deviceId = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0;
        var v = c === "x" ? r : r & 3 | 8;
        return v.toString(16);
    });
    return deviceId;
};
var getMFEName = function(airTracker2) {
    var _location_pathname_split, _location_pathname, _location;
    if (airTracker2.config.getMFEName) {
        return airTracker2.config.getMFEName();
    }
    return ((_location = location) === null || _location === void 0 ? void 0 : (_location_pathname = _location.pathname) === null || _location_pathname === void 0 ? void 0 : (_location_pathname_split = _location_pathname.split("/")) === null || _location_pathname_split === void 0 ? void 0 : _location_pathname_split[2]) || "unknown";
};
var getSessionId = function() {
    try {
        var sessionId = window.sessionStorage.getItem(SESSION_ID_STORAGE_KEY);
        if (!sessionId) {
            sessionId = generateUId();
            window.sessionStorage.setItem(SESSION_ID_STORAGE_KEY, sessionId);
        }
        return sessionId;
    } catch (e) {
        return generateUId();
    }
};
var getDeviceId = function() {
    try {
        var deviceId = window.localStorage.getItem(DEVICE_ID_STORAGE_KEY);
        if (!deviceId) {
            deviceId = generateUId();
            window.localStorage.setItem(DEVICE_ID_STORAGE_KEY, deviceId);
        }
        return deviceId;
    } catch (e) {}
};
var getRiskDeviceId = function() {
    try {
        var deviceId = window.localStorage.getItem(RISK_DEVICE_ID_STORAGE_KEY);
        return deviceId || void 0;
    } catch (e) {}
};
var getNetworkType = function() {
    var _navigator;
    var netType = "";
    var arr = navigator.userAgent.match(/NetType\/(\w+)/);
    if (arr) {
        var ref;
        ref = _sliced_to_array(arr, 2), netType = ref[1], ref;
    } else if ((_navigator = navigator) === null || _navigator === void 0 ? void 0 : _navigator.connection) {
        var _navigator_connection, _navigator1, _navigator_connection1, _navigator2;
        netType = ((_navigator1 = navigator) === null || _navigator1 === void 0 ? void 0 : (_navigator_connection = _navigator1.connection) === null || _navigator_connection === void 0 ? void 0 : _navigator_connection.effectiveType) || ((_navigator2 = navigator) === null || _navigator2 === void 0 ? void 0 : (_navigator_connection1 = _navigator2.connection) === null || _navigator_connection1 === void 0 ? void 0 : _navigator_connection1.type);
    }
    if (!netType) {
        netType = "unknown";
    }
    return parseNetType(netType);
};
var parseNetType = function(net) {
    net = String(net).toLowerCase();
    if (net.indexOf("4g") >= 0) return "net4g" /* net4g */ ;
    if (net.indexOf("wifi") >= 0) return "wifi" /* wifi */ ;
    if (net.indexOf("5g") >= 0) return "net5g" /* net5g */ ;
    if (net.indexOf("3g") >= 0) return "net3g" /* net3g */ ;
    if (net.indexOf("2g") >= 0) return "net2g" /* net2g */ ;
    return "unknown" /* unknown */ ;
};
var canUseResourceTiming = function canUseResourceTiming() {
    return typeof window.performance !== "undefined" && typeof window.performance.clearResourceTimings === "function" && typeof window.performance.getEntriesByType === "function" && typeof window.performance.now === "function";
};
var formatUrl = function(url) {
    if (typeof url === "string") {
        var hostUrl = url.split("?")[0] || "";
        return hostUrl.slice(0, 200);
    }
    return url;
};
var getQueryString = function(url) {
    if (typeof url === "string") {
        return url.split("?")[1] || "";
    }
    return url;
};
var urlIsHttps = function(url) {
    var isHostProtocol = typeof url === "string" && url.startsWith("//");
    return isHostProtocol ? typeof location !== "undefined" && location.protocol === "https:" : /^https/.test(url);
};
var getReportVal = function(rawVal, isDefaultByString) {
    if (typeof rawVal === "number") {
        return rawVal;
    }
    if (typeof rawVal === "string") {
        return rawVal;
    }
    return isDefaultByString ? "" /* string */  : -1 /* number */ ;
};
var checkIfInWhiteList = function(url, airTracker2) {
    var _airTracker2_config_assetSpeedMonitoringWhiteListByMFE;
    var isInWhiteList = false;
    if (!airTracker2.config.isWebappContainer) {
        isInWhiteList = checkIfInWhiteListHelper(url, airTracker2.config.assetSpeedMonitoringWhiteList);
        return isInWhiteList;
    }
    var MFEName = getMFEName(airTracker2);
    return isInWhiteList || checkIfInWhiteListHelper(url, (_airTracker2_config_assetSpeedMonitoringWhiteListByMFE = airTracker2.config.assetSpeedMonitoringWhiteListByMFE) === null || _airTracker2_config_assetSpeedMonitoringWhiteListByMFE === void 0 ? void 0 : _airTracker2_config_assetSpeedMonitoringWhiteListByMFE[MFEName]);
};
var checkIfInWhiteListHelper = function(url, whiteList) {
    if (!(whiteList === null || whiteList === void 0 ? void 0 : whiteList.length)) {
        return false;
    }
    for(var i = 0; i < whiteList.length; i++){
        var cur = whiteList[i];
        if (_instanceof(cur, RegExp) && cur.test(url)) {
            return true;
        }
        if (typeof cur == "string" && url.includes(cur)) {
            return true;
        }
    }
    return false;
};
var roundSubtract = function(a, b) {
    var precision = 100;
    return Math.round((a - b) * precision) / precision;
};
// src/pipes/send.ts
var sendNormalLogPipe = function(airTracker2) {
    return function(logs) {
        var _airTracker2_MFECommonDataMap, _navigator;
        var MFEName = getMFEName(airTracker2);
        var commonData = airTracker2.config.isWebappContainer ? _object_spread({}, airTracker2.commonData || {}, ((_airTracker2_MFECommonDataMap = airTracker2.MFECommonDataMap) === null || _airTracker2_MFECommonDataMap === void 0 ? void 0 : _airTracker2_MFECommonDataMap[MFEName]) || {}) : airTracker2.commonData;
        var data = Array.isArray(logs) ? logs : [
            logs
        ];
        if (data.length === 0) {
            return;
        }
        var finalLog = {
            commonData: commonData,
            data: data
        };
        var dataBody = safeStringify(finalLog);
        var sendOptions = {
            method: "post",
            url: loggingServiceUrl(airTracker2.commonData.env),
            data: dataBody
        };
        if (typeof ((_navigator = navigator) === null || _navigator === void 0 ? void 0 : _navigator.sendBeacon) === "function") {
            var success = navigator.sendBeacon(sendOptions.url, dataBody);
            if (!success) {
                request(sendOptions);
            }
        } else {
            request(sendOptions);
        }
    };
};
var sendMetricsLogPipe = function(airTracker2) {
    return function(logs) {
        var _airTracker2_MFECommonDataMap, _navigator;
        var MFEName = getMFEName(airTracker2);
        var commonData = airTracker2.config.isWebappContainer ? _object_spread({}, airTracker2.commonData || {}, ((_airTracker2_MFECommonDataMap = airTracker2.MFECommonDataMap) === null || _airTracker2_MFECommonDataMap === void 0 ? void 0 : _airTracker2_MFECommonDataMap[MFEName]) || {}) : airTracker2.commonData;
        var finalLog = {
            commonData: commonData,
            data: Array.isArray(logs) ? logs : [
                logs
            ]
        };
        var dataBody = safeStringify(finalLog);
        var sendOptions = {
            method: "post",
            url: loggingServiceMetricsUrl(airTracker2.commonData.env),
            data: dataBody
        };
        if (typeof ((_navigator = navigator) === null || _navigator === void 0 ? void 0 : _navigator.sendBeacon) === "function") {
            var success = navigator.sendBeacon(sendOptions.url, dataBody);
            if (!success) {
                request(sendOptions);
            }
        } else {
            request(sendOptions);
        }
    };
};
// src/pipes/throttle.ts
var throttlePipe = function(airTracker2, maxLength) {
    var timer;
    var retainedLogs = [];
    var config = airTracker2.config;
    window.addEventListener("beforeunload", function() {
        if (retainedLogs.length > 0) {
            airTracker2.immediateSendPipe(retainedLogs.splice(0, retainedLogs.length));
            timer && clearTimeout(timer);
        }
    });
    return function(log, resolve) {
        retainedLogs.push(log);
        airTracker2.lifeCycle.on(lifeCycleEventName.onCommonUpdated, function() {
            if (retainedLogs.length > 0) {
                resolve === null || resolve === void 0 ? void 0 : resolve(retainedLogs.splice(0, retainedLogs.length));
                airTracker2.lifeCycle.remove(lifeCycleEventName.onCommonUpdated);
                timer && clearTimeout(timer);
            }
        });
        if (maxLength && retainedLogs.length >= maxLength) {
            resolve === null || resolve === void 0 ? void 0 : resolve(retainedLogs.splice(0, retainedLogs.length));
            timer && clearTimeout(timer);
            return;
        }
        timer && clearTimeout(timer);
        timer = setTimeout(function() {
            timer = null;
            if (retainedLogs.length > 0) {
                resolve === null || resolve === void 0 ? void 0 : resolve(retainedLogs.splice(0, retainedLogs.length));
                airTracker2.lifeCycle.remove(lifeCycleEventName.onCommonUpdated);
            }
        }, config.delay);
    };
};
// src/pipes/errorRepeatLimit.ts
var errorLogLimitPipe = function(airTracker2) {
    var errorLogMap = {};
    return function(logs, resolve) {
        var maxNum = typeof airTracker2.config.errorRepeatTime === "number" ? airTracker2.config.errorRepeatTime : 5;
        if (maxNum === 0) {
            return resolve === null || resolve === void 0 ? void 0 : resolve(logs);
        }
        resolve === null || resolve === void 0 ? void 0 : resolve(logs.filter(function(log) {
            if (log.severity == "autoDetectError" /* AUTO_DETECT_ERROR */ ) {
                errorLogMap[log.error] = errorLogMap[log.error] || 0;
                errorLogMap[log.error] += 1;
                if (errorLogMap[log.error] > maxNum) {
                    return false;
                }
                return true;
            }
            return true;
        }));
    };
};
// src/pipes/index.ts
var noop = function() {};
var createPipeline = function(pipeArr) {
    if (!pipeArr || !pipeArr.reduce || !pipeArr.length) {
        throw new TypeError("createPipeline need at least one function param");
    }
    if (pipeArr.length === 1) {
        return function(msg, resolve) {
            pipeArr[0](msg, resolve || noop);
        };
    }
    return pipeArr.reduce(function(prePipe, pipe) {
        return function(msg) {
            var nextPipe = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : noop;
            return prePipe(msg, function(msg2) {
                return pipe === null || pipe === void 0 ? void 0 : pipe(msg2, nextPipe);
            });
        };
    });
};
// src/util/eventEmitter.ts
var EventEmitter = /*#__PURE__*/ function() {
    function EventEmitter() {
        var _this = this;
        _class_call_check(this, EventEmitter);
        this.emit = function(name, data) {
            if (!_this) return;
            var events = _this.eventsList[name];
            var handler;
            if (events === null || events === void 0 ? void 0 : events.length) {
                events = events.slice();
                for(var i = 0; i < events.length; i++){
                    handler = events[i];
                    try {
                        var result = handler.callback.apply(_this, [
                            data
                        ]);
                        if (1 === handler.type) {
                            _this.remove(name, handler.callback);
                        }
                        if (false === result) {
                            break;
                        }
                    } catch (e) {
                        throw e;
                    }
                }
            }
            return _this;
        };
        this.eventsList = {};
    }
    _create_class(EventEmitter, [
        {
            key: "indexOf",
            value: function indexOf(array, value) {
                for(var i = 0; i < array.length; i++){
                    if (array[i].callback === value) {
                        return i;
                    }
                }
                return -1;
            }
        },
        {
            key: "on",
            value: function on(name, callback) {
                var type = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : 0;
                if (!this) return;
                var events = this.eventsList[name];
                if (!events) {
                    this.eventsList[name] = [];
                    events = this.eventsList[name];
                }
                if (this.indexOf(events, callback) === -1) {
                    var handler = {
                        name: name,
                        type: type || 0,
                        callback: callback
                    };
                    events.push(handler);
                    return this;
                }
                return this;
            }
        },
        {
            key: "one",
            value: function one(name, callback) {
                this.on(name, callback, 1);
            }
        },
        {
            key: "remove",
            value: function remove(name, callback) {
                if (!this) return;
                var events = this.eventsList[name];
                if (!events) {
                    return null;
                }
                if (!callback) {
                    try {
                        delete this.eventsList[name];
                    } catch (e) {}
                    return null;
                }
                if (events.length) {
                    var index = this.indexOf(events, callback);
                    events.splice(index, 1);
                }
                return this;
            }
        },
        {
            key: "clear",
            value: function clear() {
                this.eventsList = {};
            }
        }
    ]);
    return EventEmitter;
}();
// src/interface/plugins.ts
var AirTrackerPlugin = /*#__PURE__*/ function() {
    function AirTrackerPlugin(pluginOption) {
        _class_call_check(this, AirTrackerPlugin);
        this.name = "";
        // initiation flag
        this.isInit = false;
        this.name = pluginOption.name;
        this.option = pluginOption;
    }
    _create_class(AirTrackerPlugin, [
        {
            key: "patch",
            value: function patch(airTracker2) {
                if (!this.isInit) {
                    var _this_option_setUp, _this_option;
                    this.isInit = true;
                    (_this_option = this.option) === null || _this_option === void 0 ? void 0 : (_this_option_setUp = _this_option.setUp) === null || _this_option_setUp === void 0 ? void 0 : _this_option_setUp.call(this.option, airTracker2);
                }
            }
        },
        {
            key: "uninstall",
            value: function uninstall() {
                var _this_option_destroy, _this_option;
                (_this_option = this.option) === null || _this_option === void 0 ? void 0 : (_this_option_destroy = _this_option.destroy) === null || _this_option_destroy === void 0 ? void 0 : _this_option_destroy.apply(this);
                this.isInit = false;
            }
        }
    ]);
    return AirTrackerPlugin;
}();
// src/plugins/onErrorPlugin.ts
var airTracker;
var onError = function(event) {
    var errString = "".concat(safeStringify(event.message) || "", " @ (").concat(safeStringify(event.filename) || "", ":").concat(event.lineno || 0, ":").concat(event.colno || 0, ")\n").concat(safeStringify(event.error || ""));
    airTracker.normalLogPipeLine({
        severity: "autoDetectError" /* AUTO_DETECT_ERROR */ ,
        eventName: "windowOnError" /* windowOnError */ ,
        extraInfo: {
            error: errString
        }
    });
};
var unhandledrejectionHandler = function(event) {
    var reason = event && safeStringify(event.reason);
    airTracker.normalLogPipeLine({
        severity: "autoDetectError" /* AUTO_DETECT_ERROR */ ,
        eventName: "promiseError" /* promiseError */ ,
        extraInfo: {
            error: reason
        }
    });
};
var errorHandler = function(event) {
    var target = (event === null || event === void 0 ? void 0 : event.target) || (event === null || event === void 0 ? void 0 : event.srcElement);
    if (!target) {
        return;
    }
    var url = (target === null || target === void 0 ? void 0 : target.src) || (target === null || target === void 0 ? void 0 : target.href);
    var tagName = target.tagName;
    var staticFileType = "unknown" /* unknown */ ;
    if (typeof url === "string" && tagName) {
        if (window.location.href.indexOf(url) > -1) {
            return;
        }
        if (/\.js$/.test(url)) {
            staticFileType = "script" /* script */ ;
        } else if (/\.css$/.test(url)) {
            staticFileType = "css" /* css */ ;
        } else {
            switch(tagName.toLowerCase()){
                case "script":
                    staticFileType = "script" /* script */ ;
                    break;
                case "link":
                    staticFileType = "css" /* css */ ;
                    break;
                case "img":
                    staticFileType = "image" /* img */ ;
                    break;
                case "audio":
                case "video":
                    staticFileType = "media" /* media */ ;
                    break;
                default:
                    return;
            }
        }
        airTracker.normalLogPipeLine({
            severity: "autoDetectError" /* AUTO_DETECT_ERROR */ ,
            eventName: "staticFileLoadError" /* staticFileLoadError */ ,
            extraInfo: {
                staticFileType: staticFileType,
                error: "".concat(tagName, " load fail: ").concat(url)
            }
        });
    }
};
var errorDetectionPlugin = new AirTrackerPlugin({
    name: "errorDetectionPlugin",
    setUp: function(airTrackerInst) {
        airTracker = airTrackerInst;
        window.addEventListener("error", onError);
        window.addEventListener("unhandledrejection", unhandledrejectionHandler);
        window.document.addEventListener("error", errorHandler, true);
    },
    destroy: function() {
        window.removeEventListener("unhandledrejection", unhandledrejectionHandler);
        window.document.removeEventListener("error", errorHandler, true);
        window.removeEventListener("error", onError);
    }
});
// src/plugins/assetsSpeedPlugin.ts
var ASSETS_INITIATOR_TYPE = [
    "img",
    "css",
    "script",
    "link",
    "audio",
    "video",
    "iframe"
];
var COLLECTED_ENTRY_TYPE = "resource";
var generateSpeedLog = function(entry) {
    var resourceURl = entry.name;
    return {
        url: formatUrl(resourceURl),
        method: "get",
        // duration = redirect + DNS + CONNECTION + request + response
        duration: Number(entry.duration.toFixed(2)),
        // time it takes to load
        type: "static",
        isHttps: urlIsHttps(resourceURl),
        urlQuery: getQueryString(resourceURl),
        domainLookup: getReportVal(entry.domainLookupEnd - entry.domainLookupStart),
        connectTime: getReportVal(entry.connectEnd - entry.connectStart)
    };
};
var publishSpeedLogs = function(entries, airTracker2) {
    for(var i = 0, l = entries.length; i < l; i++){
        var entry = entries[i];
        if (ASSETS_INITIATOR_TYPE.indexOf(entry.initiatorType) !== -1 && checkIfInWhiteList(entry.name, airTracker2)) {
            airTracker2.normalLogPipeLine({
                severity: "performance" /* PERFORMANCE */ ,
                eventName: "assets_speed" /* ASSETS_SPEED */ ,
                extraInfo: {
                    log: generateSpeedLog(entry)
                }
            });
        }
    }
};
var interval;
var observer;
var assetSpeedPlugin = new AirTrackerPlugin({
    name: "assetsSpeedPlugin",
    setUp: function(airTracker2) {
        if (!canUseResourceTiming()) return;
        var collectCur = 0;
        window.performance.onresourcetimingbufferfull = function() {
            collectCur = 0;
            window.performance.clearResourceTimings();
        };
        if (typeof window.PerformanceObserver === "function") {
            publishSpeedLogs(window.performance.getEntriesByType(COLLECTED_ENTRY_TYPE), airTracker2);
            observer = new window.PerformanceObserver(function(list) {
                publishSpeedLogs(list.getEntries(), airTracker2);
            });
            observer.observe({
                entryTypes: [
                    COLLECTED_ENTRY_TYPE
                ]
            });
        } else {
            interval = setInterval(function() {
                var allEntries = window.performance.getEntriesByType(COLLECTED_ENTRY_TYPE);
                var collectEntries = allEntries.slice(collectCur);
                collectCur = allEntries.length;
                publishSpeedLogs(collectEntries, airTracker2);
            }, 3e3);
        }
    },
    destroy: function() {
        observer === null || observer === void 0 ? void 0 : observer.disconnect();
        interval && clearInterval(interval);
    }
});
// src/plugins/onPageChangePlugin.ts
var observer2;
var onPageChangePlugin = new AirTrackerPlugin({
    name: "onPageChangePlugin",
    setUp: /*#__PURE__*/ function() {
        var _ref = _async_to_generator(function(airTracker2) {
            var _location, previousUrl, body, sendOnPageChangeEvent, config;
            return _ts_generator(this, function(_state) {
                previousUrl = (_location = location) === null || _location === void 0 ? void 0 : _location.href;
                body = document.querySelector("body");
                sendOnPageChangeEvent = function(param) {
                    var prevHref = param.prevHref;
                    var _location, _location1, _location2, _location3, _location4;
                    airTracker2.normalLogPipeLine({
                        severity: "autoDetectEvent" /* AUTO_DETECT_EVENT */ ,
                        eventName: "onPageChange" /* onPageChange */ ,
                        extraInfo: {
                            prevHref: prevHref,
                            href: ((_location = location) === null || _location === void 0 ? void 0 : _location.href) || "",
                            hostname: (_location1 = location) === null || _location1 === void 0 ? void 0 : _location1.hostname,
                            pathName: (_location2 = location) === null || _location2 === void 0 ? void 0 : _location2.pathname,
                            protocol: (_location3 = location) === null || _location3 === void 0 ? void 0 : _location3.protocol,
                            search: (_location4 = location) === null || _location4 === void 0 ? void 0 : _location4.search
                        }
                    });
                };
                sendOnPageChangeEvent({
                    prevHref: ""
                });
                observer2 = new MutationObserver(function() {
                    if (location.href !== previousUrl) {
                        var tempPreviousUrl = previousUrl;
                        previousUrl = location.href;
                        sendOnPageChangeEvent({
                            prevHref: tempPreviousUrl
                        });
                    }
                });
                config = {
                    subtree: true,
                    childList: true
                };
                observer2.observe(body || document, config);
                return [
                    2
                ];
            });
        });
        return function(airTracker2) {
            return _ref.apply(this, arguments);
        };
    }(),
    destroy: function() {
        observer2 === null || observer2 === void 0 ? void 0 : observer2.disconnect();
    }
});
// src/plugins/pageRetentionPlugin.ts
var observer3;
var onPageRetentionPlugin = new AirTrackerPlugin({
    name: "onPageRetentionPlugin",
    setUp: /*#__PURE__*/ function() {
        var _ref = _async_to_generator(function(airTracker2) {
            var _location, previousUrl, body, enterPageTimeStamp, startRecording, sendPageRetentionEvent, config;
            return _ts_generator(this, function(_state) {
                if (typeof document.visibilityState === "undefined") {
                    return [
                        2
                    ];
                }
                previousUrl = (_location = location) === null || _location === void 0 ? void 0 : _location.href;
                body = document.querySelector("body");
                startRecording = function() {
                    enterPageTimeStamp = Date.now();
                };
                startRecording();
                document.addEventListener("visibilitychange", function() {
                    if (document.visibilityState === "hidden") {
                        sendPageRetentionEvent({
                            href: location.href,
                            leftReason: "visibilityChange" /* visibilityChange */ 
                        });
                    } else if (document.visibilityState === "visible") {
                        startRecording();
                    }
                });
                window.addEventListener("beforeunload", function() {
                    sendPageRetentionEvent({
                        href: location.href,
                        sendImmediate: true,
                        leftReason: "pageClose" /* pageClose */ 
                    });
                });
                sendPageRetentionEvent = function(param) {
                    var href = param.href, sendImmediate = param.sendImmediate, leftReason = param.leftReason;
                    if (!enterPageTimeStamp) {
                        return;
                    }
                    var duration = Date.now() - enterPageTimeStamp;
                    if (duration === 0) {
                        return;
                    }
                    enterPageTimeStamp = void 0;
                    var payload = {
                        severity: "autoDetectEvent" /* AUTO_DETECT_EVENT */ ,
                        eventName: "onPageRetention" /* onPageRetention */ ,
                        extraInfo: {
                            href: href,
                            duration: duration,
                            leftReason: leftReason
                        }
                    };
                    if (sendImmediate) {
                        airTracker2.ImmediatePipeline(payload);
                        return;
                    }
                    airTracker2.normalLogPipeLine(payload);
                };
                observer3 = new MutationObserver(function() {
                    if (location.href !== previousUrl) {
                        var tempPreviousUrl = previousUrl;
                        previousUrl = location.href;
                        sendPageRetentionEvent({
                            href: tempPreviousUrl,
                            leftReason: "nextPage" /* nextPage */ 
                        });
                        startRecording();
                    }
                });
                config = {
                    subtree: true,
                    childList: true
                };
                observer3.observe(body || document, config);
                return [
                    2
                ];
            });
        });
        return function(airTracker2) {
            return _ref.apply(this, arguments);
        };
    }(),
    destroy: function() {
        observer3 === null || observer3 === void 0 ? void 0 : observer3.disconnect();
    }
});
// src/plugins/apiPerformancePlugin/utils.ts
var generateRandomId = function() {
    return Date.now().toString(36) + Math.random().toString(36).substring(2, 15);
};
var checkIfInWhiteList2 = function(url, whiteList) {
    var defaultWhiteList = [
        /^https?:\/\/[^\/]*\.airwallex\.com($|\/)/
    ];
    var defaultExclusionList = [
        /authenticate\/heartbeat/,
        /papluginlogs\/.*/,
        /airtracker\/.*/,
        /o11y(-staging|-demo)?\.airwallex\.com/,
        /\.(png|jpe?g|gif|svg|ico|webp|avif|css|js|json|woff2?|ttf|eot|map|txt|csv|xml)(\?.*)?(#.*)?$/i
    ];
    if (defaultExclusionList.some(function(item) {
        return item.test(url);
    })) {
        return false;
    }
    return checkIfInWhiteListHelper(url, whiteList || defaultWhiteList);
};
var findMainEntry = function(performanceItem) {
    var startTime = performanceItem.startTime, endTime = performanceItem.endTime, entries = performanceItem.entries;
    if (entries.length === 0) {
        return void 0;
    }
    if (entries.length === 1) {
        return entries[0];
    }
    var filteredEntries = entries.filter(function(entry) {
        return entry.startTime >= startTime && endTime && entry.responseEnd <= endTime;
    }).sort(function(a, b) {
        return a.startTime - b.startTime;
    });
    return filteredEntries === null || filteredEntries === void 0 ? void 0 : filteredEntries[0];
};
var round2Decimal = function(value) {
    return Math.round(value * 100) / 100;
};
var processPerformanceEntry = function(entry) {
    var startTime = entry.startTime, duration = entry.duration, responseStart = entry.responseStart, connectEnd = entry.connectEnd, connectStart = entry.connectStart, domainLookupEnd = entry.domainLookupEnd, domainLookupStart = entry.domainLookupStart, requestStart = entry.requestStart, secureConnectionStart = entry.secureConnectionStart, redirectEnd = entry.redirectEnd, redirectStart = entry.redirectStart, responseStatus = entry.responseStatus, responseEnd = entry.responseEnd, encodedBodySize = entry.encodedBodySize, decodedBodySize = entry.decodedBodySize, transferSize = entry.transferSize, deliveryType = entry.deliveryType, initiatorType = entry.initiatorType;
    return {
        startTime: round2Decimal(startTime),
        totalDuration: round2Decimal(duration),
        timeToFirstByte: roundSubtract(responseStart, startTime),
        tcpDuration: roundSubtract(connectEnd, connectStart),
        dnsDuration: roundSubtract(domainLookupEnd, domainLookupStart),
        tlsDuration: roundSubtract(requestStart, secureConnectionStart),
        redirectDuration: roundSubtract(redirectEnd, redirectStart),
        responseDuration: roundSubtract(responseEnd, responseStart),
        encodeResponseSize: encodedBodySize,
        decodeResponseSize: decodedBodySize,
        transferSize: transferSize,
        deliveryType: deliveryType,
        statusCode: responseStatus,
        initiatorType: initiatorType
    };
};
var getPayload = function(performanceItem, extraInfo) {
    var url = performanceItem.url, method = performanceItem.method, errorType = performanceItem.errorType, errorDetail = performanceItem.errorDetail;
    var mainEntry = findMainEntry(performanceItem);
    var _ref = new URL(url, window.location.href), href = _ref.href, protocol = _ref.protocol;
    var payloadFromPerformanceEntry = mainEntry ? processPerformanceEntry(mainEntry) : {};
    var payload = {
        severity: "autoDetectEvent" /* AUTO_DETECT_EVENT */ ,
        eventName: "httpClientMetrics" /* httpClientMetrics */ ,
        extraInfo: _object_spread(_object_spread_props(_object_spread({
            httpUrl: href,
            method: method,
            scheme: protocol
        }, errorType && {
            errorType: errorType
        }, errorDetail && {
            errorDetail: errorDetail
        }, payloadFromPerformanceEntry), {
            // use injected response as fallback
            statusCode: (payloadFromPerformanceEntry === null || payloadFromPerformanceEntry === void 0 ? void 0 : payloadFromPerformanceEntry.statusCode) || (performanceItem === null || performanceItem === void 0 ? void 0 : performanceItem.responseStatus),
            startTime: (payloadFromPerformanceEntry === null || payloadFromPerformanceEntry === void 0 ? void 0 : payloadFromPerformanceEntry.startTime) || (performanceItem === null || performanceItem === void 0 ? void 0 : performanceItem.startTime),
            initiatorType: (payloadFromPerformanceEntry === null || payloadFromPerformanceEntry === void 0 ? void 0 : payloadFromPerformanceEntry.initiatorType) || (performanceItem === null || performanceItem === void 0 ? void 0 : performanceItem.initiatorType)
        }), extraInfo)
    };
    return payload;
};
var isGraphQLError = function(responseBody) {
    return(// Has data or errors at root level
    responseBody && (typeof responseBody === "undefined" ? "undefined" : _type_of(responseBody)) === "object" && "errors" in responseBody && // If errors exist, they should be an array with proper GraphQL error shape
    Array.isArray(responseBody.errors) && responseBody.errors.every(function(error) {
        return error && (typeof error === "undefined" ? "undefined" : _type_of(error)) === "object" && "message" in error;
    }));
};
// src/plugins/apiPerformancePlugin/type.ts
var API_PERFORMANCE_QUEUE = "awx_airtracker_api_performance_queue";
var API_PERFORMANCE_PLUGIN_LOADED = "awx_airtracker_api_performance_plugin_loaded";
// src/plugins/apiPerformancePlugin/index.ts
var OBSERVER_WAIT_TIME_MS = 300;
var TIMEOUT_WAIT_TIME_MS = 1e4;
var globalObject = typeof window !== "undefined" ? window : typeof global !== "undefined" ? global : {};
globalObject[API_PERFORMANCE_QUEUE] = globalObject[API_PERFORMANCE_QUEUE] || [];
var handlePayloadAndSendLog = function(apiPerformanceItem, airTracker2, extraInfo) {
    setTimeout(function() {
        var payload = getPayload(apiPerformanceItem, extraInfo);
        airTracker2.observerPipeline(payload);
        globalObject[API_PERFORMANCE_QUEUE] = globalObject[API_PERFORMANCE_QUEUE].filter(function(item) {
            return item.id !== apiPerformanceItem.id;
        });
    }, OBSERVER_WAIT_TIME_MS);
};
var handleTimeout = function(requestId, airTracker2) {
    setTimeout(function() {
        var matchedApiPerformanceItem = globalObject[API_PERFORMANCE_QUEUE].find(function(item) {
            return item.id === requestId;
        });
        if (!matchedApiPerformanceItem) {
            return;
        }
        var payload = getPayload(matchedApiPerformanceItem, {
            exceptionType: "timeout-without-response"
        });
        airTracker2.observerPipeline(payload);
        globalObject[API_PERFORMANCE_QUEUE] = globalObject[API_PERFORMANCE_QUEUE].filter(function(item) {
            return item.id !== requestId;
        });
    }, TIMEOUT_WAIT_TIME_MS);
};
var injectXMLHttpRequest = function(airTracker2) {
    var originalOpen = XMLHttpRequest.prototype.open;
    var originalSend = XMLHttpRequest.prototype.send;
    XMLHttpRequest.prototype.open = function injectXMLHttpRequest(method, url, async, username, password) {
        try {
            var _airTracker2_config;
            var requestId = generateRandomId();
            this.__requestId = requestId;
            var urlString = _instanceof(url, URL) ? url.href : new URL(url, window.location.href).href;
            if (!checkIfInWhiteList2(urlString, (_airTracker2_config = airTracker2.config) === null || _airTracker2_config === void 0 ? void 0 : _airTracker2_config.enableRecordAPIPerformanceWhitelist)) {
                return originalOpen.call(this, method, url, async !== null && async !== void 0 ? async : true, username, password);
            }
            var apiPerformanceItem = {
                id: requestId,
                startTime: performance.now(),
                method: method,
                url: urlString,
                entries: [],
                initiatorType: "xmlhttprequest"
            };
            globalObject[API_PERFORMANCE_QUEUE].push(apiPerformanceItem);
            handleTimeout(requestId, airTracker2);
            return originalOpen.call(this, method, url, async !== null && async !== void 0 ? async : true, username, password);
        } catch (error) {
            airTracker2.normalLogPipeLine({
                severity: "warn" /* WARN */ ,
                eventName: "apiPerformancePluginError",
                extraInfo: {
                    message: "Inject XHR open error",
                    error: error
                }
            });
            return originalOpen.call(this, method, url, async !== null && async !== void 0 ? async : true, username, password);
        }
    };
    XMLHttpRequest.prototype.send = function sendXHRRequest() {
        var _this = this;
        for(var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++){
            args[_key] = arguments[_key];
        }
        try {
            var matchedApiPerformanceItem = globalObject[API_PERFORMANCE_QUEUE].find(function(item) {
                return item.id === _this.__requestId;
            });
            if (!matchedApiPerformanceItem) {
                return originalSend.apply(this, args);
            }
            var _this1 = this;
            this.addEventListener("loadend", /*#__PURE__*/ _async_to_generator(function() {
                var endTime, duration, status, responseBody, blob, parsedBody, e;
                return _ts_generator(this, function(_state) {
                    switch(_state.label){
                        case 0:
                            endTime = performance.now();
                            duration = endTime - matchedApiPerformanceItem.startTime;
                            status = _this1.status;
                            if (status === 0) {
                                return [
                                    2
                                ];
                            }
                            matchedApiPerformanceItem.endTime = endTime;
                            matchedApiPerformanceItem.duration = duration;
                            matchedApiPerformanceItem.responseStatus = status;
                            _state.label = 1;
                        case 1:
                            _state.trys.push([
                                1,
                                6,
                                ,
                                7
                            ]);
                            if (!(_this1.responseType === "" || _this1.responseType === "text")) return [
                                3,
                                2
                            ];
                            responseBody = _this1.responseText;
                            return [
                                3,
                                5
                            ];
                        case 2:
                            if (!(_this1.responseType === "json")) return [
                                3,
                                3
                            ];
                            responseBody = _this1.response;
                            return [
                                3,
                                5
                            ];
                        case 3:
                            if (!(_this1.responseType === "blob")) return [
                                3,
                                5
                            ];
                            blob = _this1.response;
                            if (!((blob === null || blob === void 0 ? void 0 : blob.type) === "application/json")) return [
                                3,
                                5
                            ];
                            return [
                                4,
                                blob.text()
                            ];
                        case 4:
                            responseBody = _state.sent();
                            _state.label = 5;
                        case 5:
                            parsedBody = responseBody && typeof responseBody === "string" ? JSON.parse(responseBody) : responseBody;
                            if (parsedBody && isGraphQLError(parsedBody)) {
                                matchedApiPerformanceItem.errorType = "graphqlError";
                                matchedApiPerformanceItem.errorDetail = parsedBody.errors;
                            }
                            return [
                                3,
                                7
                            ];
                        case 6:
                            e = _state.sent();
                            airTracker2.normalLogPipeLine({
                                severity: "warn" /* WARN */ ,
                                eventName: "apiPerformancePluginError",
                                extraInfo: {
                                    message: "Parse response body error in XHR",
                                    url: matchedApiPerformanceItem === null || matchedApiPerformanceItem === void 0 ? void 0 : matchedApiPerformanceItem.url,
                                    error: e
                                }
                            });
                            return [
                                3,
                                7
                            ];
                        case 7:
                            handlePayloadAndSendLog(matchedApiPerformanceItem, airTracker2);
                            return [
                                2
                            ];
                    }
                });
            }));
            this.addEventListener("error", function() {
                handlePayloadAndSendLog(matchedApiPerformanceItem, airTracker2, {
                    exceptionType: "network-error"
                });
            });
            this.addEventListener("abort", function() {
                handlePayloadAndSendLog(matchedApiPerformanceItem, airTracker2, {
                    exceptionType: "abort-error"
                });
            });
            this.addEventListener("timeout", function() {
                handlePayloadAndSendLog(matchedApiPerformanceItem, airTracker2, {
                    exceptionType: "timeout-error"
                });
            });
            return originalSend.apply(this, args);
        } catch (error) {
            airTracker2.normalLogPipeLine({
                severity: "warn" /* WARN */ ,
                eventName: "apiPerformancePluginError",
                extraInfo: {
                    message: "InjectXHR send error",
                    error: error
                }
            });
            throw error;
        }
    };
};
var injectFetch = function(airTracker2) {
    var originalFetch = window.fetch;
    window.fetch = /*#__PURE__*/ _async_to_generator(function() {
        var _len, args, _key, _airTracker2_config, apiPerformanceItem, url, _this, startTime, id, response, endTime, clonedResponse, responseBody, error, error1, exceptionType;
        var _arguments = arguments;
        return _ts_generator(this, function(_state) {
            switch(_state.label){
                case 0:
                    for(_len = _arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++){
                        args[_key] = _arguments[_key];
                    }
                    url = new URL(_instanceof(args[0], Request) ? args[0].url : String(args[0]), window.location.href).href;
                    if (!!checkIfInWhiteList2(url, (_airTracker2_config = airTracker2.config) === null || _airTracker2_config === void 0 ? void 0 : _airTracker2_config.enableRecordAPIPerformanceWhitelist)) return [
                        3,
                        2
                    ];
                    return [
                        4,
                        originalFetch.apply(this, args)
                    ];
                case 1:
                    return [
                        2,
                        _state.sent()
                    ];
                case 2:
                    _state.trys.push([
                        2,
                        9,
                        ,
                        10
                    ]);
                    startTime = performance.now();
                    id = generateRandomId();
                    apiPerformanceItem = {
                        id: id,
                        startTime: startTime,
                        method: ((_this = _instanceof(args[0], Request) ? args[0] : args[1]) === null || _this === void 0 ? void 0 : _this.method) || "GET",
                        url: url,
                        entries: [],
                        initiatorType: "fetch"
                    };
                    globalObject[API_PERFORMANCE_QUEUE].push(apiPerformanceItem);
                    handleTimeout(id, airTracker2);
                    return [
                        4,
                        originalFetch.apply(this, args)
                    ];
                case 3:
                    response = _state.sent();
                    endTime = performance.now();
                    apiPerformanceItem.endTime = endTime;
                    apiPerformanceItem.duration = endTime - startTime;
                    apiPerformanceItem.responseStatus = response.status;
                    clonedResponse = response.clone();
                    _state.label = 4;
                case 4:
                    _state.trys.push([
                        4,
                        7,
                        ,
                        8
                    ]);
                    if (!clonedResponse.ok) return [
                        3,
                        6
                    ];
                    return [
                        4,
                        clonedResponse.json()
                    ];
                case 5:
                    responseBody = _state.sent();
                    if (responseBody && isGraphQLError(responseBody)) {
                        apiPerformanceItem.errorType = "graphqlError";
                        apiPerformanceItem.errorDetail = responseBody.errors;
                    }
                    _state.label = 6;
                case 6:
                    return [
                        3,
                        8
                    ];
                case 7:
                    error = _state.sent();
                    airTracker2.normalLogPipeLine({
                        severity: "warn" /* WARN */ ,
                        eventName: "apiPerformancePluginError",
                        extraInfo: {
                            message: "Parse response body error in fetch",
                            url: apiPerformanceItem === null || apiPerformanceItem === void 0 ? void 0 : apiPerformanceItem.url,
                            error: error
                        }
                    });
                    return [
                        3,
                        8
                    ];
                case 8:
                    handlePayloadAndSendLog(apiPerformanceItem, airTracker2);
                    return [
                        2,
                        response
                    ];
                case 9:
                    error1 = _state.sent();
                    if (apiPerformanceItem && apiPerformanceItem.id && globalObject[API_PERFORMANCE_QUEUE].find(function(item) {
                        return item.id === (apiPerformanceItem === null || apiPerformanceItem === void 0 ? void 0 : apiPerformanceItem.id);
                    })) {
                        exceptionType = "fetch-error";
                        if (_instanceof(error1, TypeError)) {
                            exceptionType = "network-error";
                        } else if (_instanceof(error1, DOMException) && error1.name === "AbortError") {
                            exceptionType = "abort-error";
                        }
                        handlePayloadAndSendLog(apiPerformanceItem, airTracker2, {
                            exceptionType: exceptionType
                        });
                    } else {
                        airTracker2.normalLogPipeLine({
                            severity: "warn" /* WARN */ ,
                            eventName: "apiPerformancePluginError",
                            extraInfo: {
                                message: "Inject fetch error",
                                error: error1
                            }
                        });
                    }
                    throw error1;
                case 10:
                    return [
                        2
                    ];
            }
        });
    });
};
var setUpAPIPerformancePlugin = function(airTracker2) {
    var _window;
    injectXMLHttpRequest(airTracker2);
    injectFetch(airTracker2);
    if ((_window = window) === null || _window === void 0 ? void 0 : _window.PerformanceObserver) {
        var observer4 = new window.PerformanceObserver(function(list) {
            var entries = list.getEntries();
            entries.forEach(function(entry) {
                var resourceEntry = entry;
                if (resourceEntry.initiatorType === "fetch" || resourceEntry.initiatorType === "xmlhttprequest") {
                    var items = globalObject[API_PERFORMANCE_QUEUE].filter(function(item) {
                        return item.url === resourceEntry.name;
                    });
                    items.forEach(function(item) {
                        item.entries.push(resourceEntry);
                    });
                }
            });
        });
        observer4.observe({
            entryTypes: [
                "resource"
            ],
            buffered: true
        });
    } else {
        airTracker2.normalLogPipeLine({
            severity: "warn" /* WARN */ ,
            eventName: "apiPerformancePluginError",
            extraInfo: {
                error: "PerformanceObserver Not Supported"
            }
        });
    }
};
var apiPerformancePlugin = new AirTrackerPlugin({
    name: "apiPerformancePlugin",
    setUp: /*#__PURE__*/ function() {
        var _ref = _async_to_generator(function(airTracker2) {
            var _airTracker2_config, _airTracker2_config1;
            return _ts_generator(this, function(_state) {
                if (globalObject[API_PERFORMANCE_PLUGIN_LOADED]) {
                    return [
                        2
                    ];
                }
                if (((_airTracker2_config = airTracker2.config) === null || _airTracker2_config === void 0 ? void 0 : _airTracker2_config.enableRecordAPIPerformance) && Math.random() < (((_airTracker2_config1 = airTracker2.config) === null || _airTracker2_config1 === void 0 ? void 0 : _airTracker2_config1.enableRecordAPIPerformancePercentage) || 0)) {
                    globalObject[API_PERFORMANCE_PLUGIN_LOADED] = true;
                    setUpAPIPerformancePlugin(airTracker2);
                }
                return [
                    2
                ];
            });
        });
        return function(airTracker2) {
            return _ref.apply(this, arguments);
        };
    }()
});
// src/util/version.ts
var SDK_VERSION = "3.1.0";
// src/airTracker.ts
var AirTracker = /*#__PURE__*/ function() {
    function AirTracker(param) {
        var _this = this;
        var config = param.config, _param_plugins = param.plugins, plugins = _param_plugins === void 0 ? [] : _param_plugins;
        _class_call_check(this, AirTracker);
        this.config = DEFAULT_CONFIG;
        this.lifeCycle = new EventEmitter();
        this.plugins = [];
        this._commonData = DEFAULT_COMMON_DATA;
        this._MFECommonDataMap = {};
        // customized speed log
        this.timeMap = {};
        this.packedDataConvert = function(param) {
            var severity = param.severity, eventName = param.eventName, extraInfo = param.extraInfo;
            var _location, _location1;
            var packedData = _object_spread({
                severity: severity,
                eventName: eventName,
                currentHref: _this.config.hideSearchParams ? (((_location = location) === null || _location === void 0 ? void 0 : _location.href) || "").split("?")[0] || "" : ((_location1 = location) === null || _location1 === void 0 ? void 0 : _location1.href) || ""
            }, extraInfo);
            if (_this.config.isWebappContainer) {
                packedData.MFEName = getMFEName(_this);
            }
            return packedData;
        };
        this.normalPipelineObj = createPipeline([
            throttlePipe(this, 8),
            errorLogLimitPipe(this),
            sendNormalLogPipe(this)
        ]);
        this.observerPipelineObj = createPipeline([
            throttlePipe(this, 8),
            sendMetricsLogPipe(this)
        ]);
        this.immediateSendPipe = sendNormalLogPipe(this);
        this.observerPipeline = function(param) {
            var severity = param.severity, eventName = param.eventName, extraInfo = param.extraInfo;
            var packedData = _this.packedDataConvert({
                severity: severity,
                eventName: eventName,
                extraInfo: extraInfo
            });
            return _this.observerPipelineObj(packedData);
        };
        this.normalLogPipeLine = function(param) {
            var severity = param.severity, eventName = param.eventName, extraInfo = param.extraInfo;
            var packedData = _this.packedDataConvert({
                severity: severity,
                eventName: eventName,
                extraInfo: extraInfo
            });
            if (_this.config.disableThrottle) {
                return _this.immediateSendPipe(packedData);
            } else {
                return _this.normalPipelineObj(packedData);
            }
        };
        this.ImmediatePipeline = function(param) {
            var severity = param.severity, eventName = param.eventName, extraInfo = param.extraInfo;
            var packedData = _this.packedDataConvert({
                severity: severity,
                eventName: eventName,
                extraInfo: extraInfo
            });
            return _this.immediateSendPipe(packedData);
        };
        this.plugins = _to_consumable_array(plugins);
        this.setConfig(config);
        this.initCommonData(config);
        this.lifeCycle.emit(lifeCycleEventName.onInit);
        if (config.enableErrorMonitoring) {
            this.plugins.push(errorDetectionPlugin);
        }
        if ((config === null || config === void 0 ? void 0 : config.assetSpeedMonitoringWhiteList) && config.assetSpeedMonitoringWhiteList.length > 0 || config.isWebappContainer) {
            this.plugins.push(assetSpeedPlugin);
        }
        if (config === null || config === void 0 ? void 0 : config.enableDetectPageChange) {
            this.plugins.push(onPageChangePlugin);
        }
        if (config === null || config === void 0 ? void 0 : config.enableRecordPageRetention) {
            this.plugins.push(onPageRetentionPlugin);
        }
        if ((config === null || config === void 0 ? void 0 : config.enableRecordAPIPerformance) && (config === null || config === void 0 ? void 0 : config.enableRecordAPIPerformancePercentage) && config.enableRecordAPIPerformancePercentage > 0 && config.enableRecordAPIPerformancePercentage <= 1) {
            this.plugins.push(apiPerformancePlugin);
        }
        this.installPlugins();
    }
    _create_class(AirTracker, [
        {
            key: "installPlugins",
            value: function installPlugins() {
                var _this = this;
                this.plugins.forEach(function(item) {
                    item.patch(_this);
                });
            }
        },
        {
            key: "commonData",
            get: function get() {
                return this._commonData;
            }
        },
        {
            key: "MFECommonDataMap",
            get: function get() {
                return this._MFECommonDataMap;
            }
        },
        {
            key: "initCommonData",
            value: function initCommonData(config) {
                this._commonData.sessionId = config.useSessionStorageForSessionId ? getSessionId() : generateUId();
                this._commonData.deviceId = getDeviceId() || UNKNOWN_VALUE;
                this._commonData.riskDeviceId = getRiskDeviceId();
                this._commonData.networkType = getNetworkType() || "unknown" /* unknown */ ;
                this._commonData.env = config.env || this._commonData.env;
                this._commonData.accountId = config.accountId || UNKNOWN_VALUE;
                this._commonData.appVersion = config.appVersion;
                this._commonData.appName = config.appName;
                this._commonData.sdkVersion = SDK_VERSION;
                this.lifeCycle.emit(lifeCycleEventName.onCommonDataInit, this.commonData);
            }
        },
        {
            key: "updateCommonDataBasedOnConfig",
            value: function updateCommonDataBasedOnConfig(config) {
                this.lifeCycle.emit(lifeCycleEventName.onCommonUpdated, this._commonData);
                this._commonData.env = config.env || this._commonData.env;
                this._commonData.appName = config.appName || this._commonData.appName;
                this._commonData.env = config.env || this._commonData.env;
                this._commonData.accountId = config.accountId || this._commonData.accountId;
                this._commonData.appVersion = config.appVersion || this._commonData.appVersion;
            }
        },
        {
            // Can add extra common data based on developers' needs
            key: "updateCommonData",
            value: function updateCommonData(extraData) {
                this.lifeCycle.emit(lifeCycleEventName.onCommonUpdated, this._commonData);
                this._commonData = _object_spread({}, this._commonData, extraData);
            }
        },
        {
            key: "removeCommonDataByKey",
            value: function removeCommonDataByKey(keys) {
                var _this = this;
                this.lifeCycle.emit(lifeCycleEventName.onCommonUpdated, this._commonData);
                setTimeout(function() {
                    keys.forEach(function(item) {
                        delete _this._commonData[item];
                    });
                }, 1e3);
            }
        },
        {
            // only worked if isWebappContainer is true
            key: "addToMFECommonData",
            value: function addToMFECommonData(param) {
                var MFEName = param.MFEName, MFECommonData = param.MFECommonData;
                var _this__MFECommonDataMap;
                if (!this.config.isWebappContainer) {
                    return;
                }
                if (!MFEName || !MFECommonData) {
                    return;
                }
                this._MFECommonDataMap = _object_spread_props(_object_spread({}, this._MFECommonDataMap || {}), _define_property({}, MFEName, _object_spread({}, ((_this__MFECommonDataMap = this._MFECommonDataMap) === null || _this__MFECommonDataMap === void 0 ? void 0 : _this__MFECommonDataMap[MFEName]) || {}, MFECommonData)));
            }
        },
        {
            // support set config later
            key: "setConfig",
            value: function setConfig(config) {
                var _this = this;
                if (this.config.isWebappContainer) {
                    return;
                }
                var setValue = function(key, value) {
                    _this.config[key] = value;
                };
                Object.entries(config).forEach(function(item) {
                    var _item = _sliced_to_array(item, 2), key = _item[0], val = _item[1];
                    if ((typeof val === "undefined" ? "undefined" : _type_of(val)) !== void 0) {
                        setValue(key, val);
                    }
                });
                this.lifeCycle.emit(lifeCycleEventName.onConfigInit, this.config);
                this.updateCommonDataBasedOnConfig(config);
            }
        },
        {
            // send verbose level log
            key: "verbose",
            value: function verbose(eventName, extraInfo) {
                this.normalLogPipeLine({
                    severity: "verbose" /* VERBOSE */ ,
                    eventName: eventName,
                    extraInfo: extraInfo
                });
            }
        },
        {
            // send HTTP request log
            key: "httpLog",
            value: function httpLog(eventName, extraInfo) {
                this.normalLogPipeLine({
                    severity: "http" /* HTTP */ ,
                    eventName: eventName,
                    extraInfo: extraInfo
                });
            }
        },
        {
            // send info level log
            key: "info",
            value: function info(eventName, extraInfo) {
                this.normalLogPipeLine({
                    severity: "info" /* INFO */ ,
                    eventName: eventName,
                    extraInfo: extraInfo
                });
            }
        },
        {
            // send warn level log
            key: "warn",
            value: function warn(eventName, extraInfo) {
                this.normalLogPipeLine({
                    severity: "warn" /* WARN */ ,
                    eventName: eventName,
                    extraInfo: extraInfo
                });
            }
        },
        {
            // send error level log
            key: "error",
            value: function error(eventName, extraInfo) {
                console.error("error", eventName);
                this.normalLogPipeLine({
                    severity: "error" /* ERROR */ ,
                    eventName: eventName,
                    extraInfo: extraInfo
                });
            }
        },
        {
            // send business related log, usually used for data analysis
            // all the logs with isBusinessData == true, will be sync to bigQuery / looker
            key: "businessLog",
            value: function businessLog(eventName, extraInfo) {
                this.normalLogPipeLine({
                    severity: "info" /* INFO */ ,
                    eventName: eventName,
                    extraInfo: _object_spread_props(_object_spread({}, extraInfo), {
                        isBusinessLog: true
                    })
                });
            }
        },
        {
            // duration: milliseconds
            key: "reportDuration",
            value: function reportDuration(eventName, duration, extraInfo) {
                if (typeof eventName !== "string") {
                    console.warn("reportDuration: eventName (first param) must be a string");
                    return;
                }
                if (typeof duration !== "number") {
                    console.warn("reportDuration: duration (second param) must be number");
                    return;
                }
                if (duration < 0 || duration > 6e4) {
                    console.warn("reportDuration: duration (second param) must between 0 and 60000");
                    return;
                }
                this.normalLogPipeLine({
                    severity: "speed" /* SPEED */ ,
                    eventName: eventName,
                    extraInfo: _object_spread({
                        duration: duration
                    }, extraInfo)
                });
            }
        },
        {
            key: "getTimerKey",
            value: function getTimerKey(eventName) {
                var MFEName = getMFEName(this);
                return this.config.isWebappContainer ? "".concat(MFEName, "_").concat(eventName) : eventName;
            }
        },
        {
            key: "timeStart",
            value: function timeStart(eventName) {
                if (typeof eventName !== "string") {
                    console.warn("time: first param must be a string");
                    return;
                }
                if (this.timeMap[this.getTimerKey(eventName)]) {
                    console.warn("Timer ".concat(eventName, " already exists"));
                }
                this.timeMap[this.getTimerKey(eventName)] = Date.now();
            }
        },
        {
            key: "timeEnd",
            value: function timeEnd(eventName, delta) {
                if (typeof eventName !== "string") {
                    console.warn("timeEnd: first param must be a string");
                    return;
                }
                if (this.timeMap[this.getTimerKey(eventName)]) {
                    var duration = Date.now() - this.timeMap[this.getTimerKey(eventName)] + (delta || 0);
                    this.normalLogPipeLine({
                        severity: "speed" /* SPEED */ ,
                        eventName: eventName,
                        extraInfo: {
                            duration: duration
                        }
                    });
                    delete this.timeMap[this.getTimerKey(eventName)];
                } else {
                    console.warn("Timer key :".concat(eventName, " does not exist"));
                }
            }
        },
        {
            // only worked if isWebappContainer is true
            key: "addToAssetSpeedWhiteListByMFE",
            value: function addToAssetSpeedWhiteListByMFE(param) {
                var MFEName = param.MFEName, whiteList = param.whiteList;
                var _this_config_assetSpeedMonitoringWhiteListByMFE;
                if (!this.config.isWebappContainer) {
                    return;
                }
                if (!MFEName || !(whiteList === null || whiteList === void 0 ? void 0 : whiteList.length)) {
                    return;
                }
                if (!this.config.assetSpeedMonitoringWhiteListByMFE) {
                    this.config.assetSpeedMonitoringWhiteListByMFE = {};
                }
                if (!((_this_config_assetSpeedMonitoringWhiteListByMFE = this.config.assetSpeedMonitoringWhiteListByMFE) === null || _this_config_assetSpeedMonitoringWhiteListByMFE === void 0 ? void 0 : _this_config_assetSpeedMonitoringWhiteListByMFE[MFEName])) {
                    this.config.assetSpeedMonitoringWhiteListByMFE = _object_spread_props(_object_spread({}, this.config.assetSpeedMonitoringWhiteListByMFE), _define_property({}, MFEName, whiteList));
                } else {
                    var _this_config_assetSpeedMonitoringWhiteListByMFE_MFEName;
                    var _this_config_assetSpeedMonitoringWhiteListByMFE_MFEName1, _this_config_assetSpeedMonitoringWhiteListByMFE1;
                    (_this_config_assetSpeedMonitoringWhiteListByMFE1 = this.config.assetSpeedMonitoringWhiteListByMFE) === null || _this_config_assetSpeedMonitoringWhiteListByMFE1 === void 0 ? void 0 : (_this_config_assetSpeedMonitoringWhiteListByMFE_MFEName1 = _this_config_assetSpeedMonitoringWhiteListByMFE1[MFEName]) === null || _this_config_assetSpeedMonitoringWhiteListByMFE_MFEName1 === void 0 ? void 0 : (_this_config_assetSpeedMonitoringWhiteListByMFE_MFEName = _this_config_assetSpeedMonitoringWhiteListByMFE_MFEName1).push.apply(_this_config_assetSpeedMonitoringWhiteListByMFE_MFEName, _to_consumable_array(whiteList));
                }
            }
        },
        {
            // please call destroy from your app root component when your app unmount
            key: "destroy",
            value: function destroy() {
                if (this.config.isWebappContainer) {
                    return;
                }
                this.plugins.forEach(function(item) {
                    item.uninstall();
                });
            }
        }
    ]);
    return AirTracker;
}();
// src/index.ts
var src_default = AirTracker;
