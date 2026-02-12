/**
 * CrossDev Preload Script Template
 *
 * Use this as a custom preload: set htmlLoading.preloadPath to "crossdev_preload.js"
 * in options.json. Must expose CrossDev.invoke() and CrossDev.events for MessageRouter.
 *
 * Bridge is non-overridable: Object.defineProperty + Object.freeze so the page cannot
 * replace or delete window.CrossDev.
 *
 * Binary support: Pass ArrayBuffer/Uint8Array in payload - auto base64. Use
 * invoke(type, payload, { binaryResponse: true }) to decode result.data to ArrayBuffer.
 */
(function() {
    var _pending = new Map();
    var _eventListeners = {};

    function _ab2b64(ab) {
        var u8 = ab instanceof Uint8Array ? ab : new Uint8Array(ab);
        var bin = '';
        for (var i = 0; i < u8.length; i++) bin += String.fromCharCode(u8[i]);
        return btoa(bin);
    }
    function _b642ab(s) {
        var bin = atob(s);
        var u8 = new Uint8Array(bin.length);
        for (var i = 0; i < bin.length; i++) u8[i] = bin.charCodeAt(i);
        return u8.buffer;
    }
    function _toWire(obj) {
        if (obj === null || typeof obj !== 'object') return obj;
        if (obj instanceof ArrayBuffer) return { __base64: _ab2b64(obj) };
        if (ArrayBuffer.isView && ArrayBuffer.isView(obj))
            return { __base64: _ab2b64(obj.buffer.slice(obj.byteOffset, obj.byteOffset + obj.byteLength)) };
        if (Array.isArray(obj)) return obj.map(_toWire);
        var out = {};
        for (var k in obj) if (Object.prototype.hasOwnProperty.call(obj, k)) out[k] = _toWire(obj[k]);
        return out;
    }

    window.addEventListener('message', function(e) {
        var d = e.data;
        if (!d) return;
        if (d.type === 'crossdev:event') {
            var name = d.name, payload = d.payload || {};
            var list = _eventListeners[name];
            if (list) list.forEach(function(fn) { try { fn(payload); } catch (err) { console.error(err); } });
            return;
        }
        if (d.requestId) {
            var h = _pending.get(d.requestId);
            if (h) {
                _pending.delete(d.requestId);
                var res = d.result;
                if (h.binary && res && typeof res.data === 'string') {
                    res = Object.assign({}, res);
                    res.data = _b642ab(res.data);
                }
                d.error ? h.reject(new Error(d.error)) : h.resolve(res);
            }
        }
    });

    function _post(msg) {
        if (window.webkit && window.webkit.messageHandlers && window.webkit.messageHandlers.nativeMessage) {
            window.webkit.messageHandlers.nativeMessage.postMessage(msg);
        }
    }

    var CrossDev = {
        invoke: function(type, payload, opts) {
            var opt = opts || {};
            return new Promise(function(resolve, reject) {
                var rid = Date.now() + '-' + Math.random();
                _pending.set(rid, { resolve: resolve, reject: reject, binary: !!opt.binaryResponse });
                setTimeout(function() {
                    if (_pending.has(rid)) { _pending.delete(rid); reject(new Error('Request timeout')); }
                }, 10000);
                _post({ type: type, payload: _toWire(payload || {}), requestId: rid });
            });
        },
        events: {
            on: function(name, fn) {
                if (!_eventListeners[name]) _eventListeners[name] = [];
                _eventListeners[name].push(fn);
                return function() {
                    var i = _eventListeners[name].indexOf(fn);
                    if (i >= 0) _eventListeners[name].splice(i, 1);
                };
            }
        }
    };
    Object.freeze(CrossDev.events);
    Object.freeze(CrossDev);
    try { Object.defineProperty(window, 'CrossDev', { value: CrossDev, configurable: false, writable: false }); } catch (_) { window.CrossDev = CrossDev; }

    window.chrome = window.chrome || {};
    window.chrome.webview = window.chrome.webview || {};
    window.chrome.webview.postMessage = function(m) {
        var msg = typeof m === 'string' ? JSON.parse(m) : m;
        _post(msg);
    };
})();
