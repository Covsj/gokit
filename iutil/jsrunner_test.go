package iutil

import (
	"testing"
	"time"

	"github.com/Covsj/gokit/ilog"
)

func TestJS(t *testing.T) {
	runner := DefaultJS()
	fStr := `
	function r(e) {
        return null != e && "[object Object]" == Object.prototype.toString.call(e)
    }
	function n() {
        if ("function" == typeof Uint32Array) {
            var e = "";
            if ("undefined" != typeof crypto ? e = crypto : "undefined" != typeof msCrypto && (e = msCrypto),
            r(e) && e.getRandomValues) {
                var t = new Uint32Array(1)
                  , n = e.getRandomValues(t)[0]
                  , i = Math.pow(2, 32);
                return n / i
            }
        }
        return Qi(1e19) / 1e19
    }

	Qi = function() {
        function e() {
            return r = (9301 * r + 49297) % 233280,
            r / 233280
        }
        var t = new Date
          , r = t.getTime();
        return function(t) {
            return Math.ceil(e() * t)
        }
    }();
	function getTrackId(){
     return Number(String(n()).slice(2, 5) + String(n()).slice(2, 4) + String((new Date).getTime()).slice(-4))
	}`
	res, err := runner.CallFunctionNamed("getTrackId", fStr, []any{}, 5*time.Second)
	ilog.Info("js返回", "结果", res, "错误", err)
}
