package mutate

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yaklang/yaklang/common/jsonpath"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"github.com/yaklang/yaklang/common/utils/mixer"
	"github.com/yaklang/yaklang/common/yak/cartesian"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
)

func isBase64JSON(raw string) (string, bool) {
	decoded, err := codec.DecodeBase64Url(raw)
	if err != nil {
		return raw, false
	}
	return utils.IsJSON(string(decoded))
}

func isBase64(raw string) (string, bool) {
	decoded, err := codec.DecodeBase64Url(raw)
	if err != nil {
		return raw, false
	}
	return string(decoded), true
}

func (f *FuzzHTTPRequest) fuzzGetBase64Params(key interface{}, value interface{}) ([]*http.Request, error) {
	req, err := f.GetOriginHTTPRequest()
	if err != nil {
		return nil, err
	}
	vals := req.URL.Query()
	if vals == nil {
		vals = make(url.Values)
	}

	keys, values := InterfaceToFuzzResults(key), InterfaceToFuzzResults(value)
	if len(keys) <= 0 || len(values) <= 0 {
		return nil, utils.Errorf("GetQuery key or Values are empty...")
	}
	mix, err := mixer.NewMixer(keys, values)
	if err != nil {
		return nil, err
	}

	var reqs []*http.Request
	for {
		pairs := mix.Value()
		key, value := pairs[0], codec.EncodeBase64(pairs[1])
		req.RequestURI = ""
		newVals, err := deepCopyUrlValues(vals)
		if err != nil {
			continue
		}
		newVals.Set(key, value)
		req.URL.RawQuery = newVals.Encode()

		_req, err := rebuildHTTPRequest(req, 0)
		if err != nil {
			continue
		}
		req.URL.RawQuery = vals.Encode()
		reqs = append(reqs, _req)

		err = mix.Next()
		if err != nil {
			break
		}
	}
	return reqs, nil
}

func (f *FuzzHTTPRequest) fuzzPostBase64JsonPath(key any, jsonPath string, val any) ([]*http.Request, error) {
	req, err := f.GetOriginHTTPRequest()
	if err != nil {
		return nil, err
	}

	keyStr := utils.InterfaceToString(key)
	vals, err := url.ParseQuery(string(f.GetBody()))
	if err != nil {
		return nil, utils.Errorf("url.ParseQuery: %s", err)
	}
	originValue := vals.Get(keyStr)
	if strings.Contains(originValue, "%") {
		unescaped, err := url.QueryUnescape(originValue)
		if err == nil {
			originValue = unescaped
		}
	}
	if ret, ok := isBase64JSON(originValue); !ok {
		return nil, utils.Errorf("invalid base64 json: %s", ret)
	} else {
		originValue = ret
	}

	var reqs []*http.Request
	err = cartesian.ProductEx([][]string{
		{keyStr}, InterfaceToFuzzResults(val),
	}, func(result []string) error {
		value := result[1]
		var replaced = valueToJsonValue(value)
		for _, i := range replaced {
			_req := lowhttp.CopyRequest(req)
			originVals := make(url.Values)
			for k, v := range vals {
				if k == keyStr {
					originVals.Set(
						k,
						codec.EncodeBase64(jsonpath.ReplaceString(originValue, jsonPath, i)))
				} else {
					originVals[k] = v
				}
			}
			_req.Body = io.NopCloser(bytes.NewBufferString(originVals.Encode()))
			reqs = append(reqs, _req)
		}
		return nil
	})
	if err != nil {
		return nil, utils.Errorf("cartesian.ProductEx: %s", err)
	}
	return reqs, nil
}

func (f *FuzzHTTPRequest) fuzzGetBase64JsonPath(key any, jsonPath string, val any) ([]*http.Request, error) {
	req, err := f.GetOriginHTTPRequest()
	if err != nil {
		return nil, err
	}

	keyStr := utils.InterfaceToString(key)
	vals := req.URL.Query()
	originValue := vals.Get(keyStr)
	if strings.Contains(originValue, "%") {
		unescaped, err := url.QueryUnescape(originValue)
		if err == nil {
			originValue = unescaped
		}
	}
	if ret, ok := isBase64JSON(originValue); !ok {
		return nil, utils.Errorf("invalid base64 json: %s", ret)
	} else {
		originValue = ret
	}

	var reqs []*http.Request
	err = cartesian.ProductEx([][]string{
		{keyStr}, InterfaceToFuzzResults(val),
	}, func(result []string) error {
		value := result[1]
		var replaced = valueToJsonValue(value)
		for _, i := range replaced {
			_req := lowhttp.CopyRequest(req)
			newVals := _req.URL.Query()
			newVals.Set(keyStr, codec.EncodeBase64(jsonpath.ReplaceString(originValue, jsonPath, i)))
			_req.URL.RawQuery = newVals.Encode()
			_req.RequestURI = _req.URL.RequestURI()
			reqs = append(reqs, _req)
		}
		return nil
	})
	if err != nil {
		return nil, utils.Errorf("cartesian.ProductEx: %s", err)
	}
	return reqs, nil
}

func (f *FuzzHTTPRequest) toFuzzHTTPRequestBatch() *FuzzHTTPRequestBatch {
	return &FuzzHTTPRequestBatch{fallback: f, originRequest: f, noAutoEncode: f.noAutoEncode}
}

func (f *FuzzHTTPRequest) FuzzGetBase64Params(key, val any) FuzzHTTPRequestIf {
	reqs, err := f.fuzzGetParams(key, val, codec.EncodeBase64)
	if err != nil {
		return f.toFuzzHTTPRequestBatch()
	}
	return NewFuzzHTTPRequestBatch(f, reqs...)
}

func (f *FuzzHTTPRequest) FuzzPostBase64Params(key, val any) FuzzHTTPRequestIf {
	reqs, err := f.fuzzPostParams(key, val, codec.EncodeBase64)
	if err != nil {
		return f.toFuzzHTTPRequestBatch()
	}
	return NewFuzzHTTPRequestBatch(f, reqs...)
}

func (f *FuzzHTTPRequest) FuzzCookieBase64(key, val any) FuzzHTTPRequestIf {
	reqs, err := f.fuzzCookie(key, val, codec.EncodeBase64)
	if err != nil {
		return f.toFuzzHTTPRequestBatch()
	}
	return NewFuzzHTTPRequestBatch(f, reqs...)
}

func (f *FuzzHTTPRequest) FuzzGetBase64JsonPath(key any, jsonPath string, val any) FuzzHTTPRequestIf {
	reqs, err := f.fuzzGetBase64JsonPath(key, jsonPath, val)
	if err != nil {
		return f.toFuzzHTTPRequestBatch()
	}
	return NewFuzzHTTPRequestBatch(f, reqs...)
}

func (f *FuzzHTTPRequest) FuzzPostBase64JsonPath(key any, jsonPath string, val any) FuzzHTTPRequestIf {
	reqs, err := f.fuzzPostBase64JsonPath(key, jsonPath, val)
	if err != nil {
		return f.toFuzzHTTPRequestBatch()
	}
	return NewFuzzHTTPRequestBatch(f, reqs...)
}

func (f *FuzzHTTPRequestBatch) FuzzPostBase64JsonPath(key any, jsonPath string, val any) FuzzHTTPRequestIf {
	if len(f.nextFuzzRequests) <= 0 {
		return f.fallback.FuzzPostBase64JsonPath(key, jsonPath, val)
	}

	var reqs []FuzzHTTPRequestIf
	for _, req := range f.nextFuzzRequests {
		reqs = append(reqs, req.FuzzPostBase64JsonPath(key, jsonPath, val))
	}

	return f.toFuzzHTTPRequestIf(reqs)
}

func (f *FuzzHTTPRequestBatch) FuzzGetBase64JsonPath(key any, jsonPath string, val any) FuzzHTTPRequestIf {
	if len(f.nextFuzzRequests) <= 0 {
		return f.fallback.FuzzGetBase64JsonPath(key, jsonPath, val)
	}

	var reqs []FuzzHTTPRequestIf
	for _, req := range f.nextFuzzRequests {
		reqs = append(reqs, req.FuzzGetBase64JsonPath(key, jsonPath, val))
	}

	return f.toFuzzHTTPRequestIf(reqs)
}
