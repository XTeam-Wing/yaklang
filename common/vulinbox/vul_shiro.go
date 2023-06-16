package vulinbox

import (
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yso"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var keyList = []string{
	"kPH+bIxk5D2deZiIxcaaaA==",
	"4AvVhmFLUs0KTA3Kprsdag==",
	"Z3VucwAAAAAAAAAAAAAAAA==",
	"fCq+/xW488hMTCD+cmJ3aQ==",
	"0AvVhmFLUs0KTA3Kprsdag==",
	"1AvVhdsgUs0FSA3SDFAdag==",
	"1QWLxg+NYmxraMoxAXu/Iw==",
	"25BsmdYwjnfcWmnhAciDDg==",
	"2AvVhdsgUs0FSA3SDFAdag==",
	"3AvVhmFLUs0KTA3Kprsdag==",
	"3JvYhmBLUs0ETA5Kprsdag==",
	"r0e3c16IdVkouZgk1TKVMg==",
	"5aaC5qKm5oqA5pyvAAAAAA==",
	"5AvVhmFLUs0KTA3Kprsdag==",
	"6AvVhmFLUs0KTA3Kprsdag==",
	"6NfXkC7YVCV5DASIrEm1Rg==",
	"6ZmI6I2j5Y+R5aSn5ZOlAA==",
	"cmVtZW1iZXJNZQAAAAAAAA==",
	"7AvVhmFLUs0KTA3Kprsdag==",
	"8AvVhmFLUs0KTA3Kprsdag==",
	"8BvVhmFLUs0KTA3Kprsdag==",
	"9AvVhmFLUs0KTA3Kprsdag==",
	"OUHYQzxQ/W9e/UjiAGu6rg==",
	"a3dvbmcAAAAAAAAAAAAAAA==",
	"aU1pcmFjbGVpTWlyYWNsZQ==",
	"bWljcm9zAAAAAAAAAAAAAA==",
	"bWluZS1hc3NldC1rZXk6QQ==",
	"bXRvbnMAAAAAAAAAAAAAAA==",
	"ZUdsaGJuSmxibVI2ZHc9PQ==",
	"wGiHplamyXlVB11UXWol8g==",
	"U3ByaW5nQmxhZGUAAAAAAA==",
	"MTIzNDU2Nzg5MGFiY2RlZg==",
	"L7RioUULEFhRyxM7a2R/Yg==",
	"a2VlcE9uR29pbmdBbmRGaQ==",
	"WcfHGU25gNnTxTlmJMeSpw==",
	"OY//C4rhfwNxCQAQCrQQ1Q==",
	"5J7bIJIV0LQSN3c9LPitBQ==",
	"f/SY5TIve5WWzT4aQlABJA==",
	"bya2HkYo57u6fWh5theAWw==",
	"WuB+y2gcHRnY2Lg9+Aqmqg==",
	"kPv59vyqzj00x11LXJZTjJ2UHW48jzHN",
	"3qDVdLawoIr1xFd6ietnwg==",
	"ZWvohmPdUsAWT3=KpPqda",
	"YI1+nBV//m7ELrIyDHm6DQ==",
	"6Zm+6I2j5Y+R5aS+5ZOlAA==",
	"2A2V+RFLUs+eTA3Kpr+dag==",
	"6ZmI6I2j3Y+R1aSn5BOlAA==",
	"SkZpbmFsQmxhZGUAAAAAAA==",
	"2cVtiE83c4lIrELJwKGJUw==",
	"fsHspZw/92PrS3XrPW+vxw==",
	"XTx6CKLo/SdSgub+OPHSrw==",
	"sHdIjUN6tzhl8xZMG3ULCQ==",
	"O4pdf+7e+mZe8NyxMTPJmQ==",
	"HWrBltGvEZc14h9VpMvZWw==",
	"rPNqM6uKFCyaL10AK51UkQ==",
	"Y1JxNSPXVwMkyvES/kJGeQ==",
	"lT2UvDUmQwewm6mMoiw4Ig==",
	"MPdCMZ9urzEA50JDlDYYDg==",
	"xVmmoltfpb8tTceuT5R7Bw==",
	"c+3hFGPjbgzGdrC+MHgoRQ==",
	"ClLk69oNcA3m+s0jIMIkpg==",
	"Bf7MfkNR0axGGptozrebag==",
	"1tC/xrDYs8ey+sa3emtiYw==",
	"ZmFsYWRvLnh5ei5zaGlybw==",
	"cGhyYWNrY3RmREUhfiMkZA==",
	"IduElDUpDDXE677ZkhhKnQ==",
	"yeAAo1E8BOeAYfBlm4NG9Q==",
	"cGljYXMAAAAAAAAAAAAAAA==",
	"2itfW92XazYRi5ltW0M2yA==",
	"XgGkgqGqYrix9lI6vxcrRw==",
	"ertVhmFLUs0KTA3Kprsdag==",
	"5AvVhmFLUS0ATA4Kprsdag==",
	"s0KTA3mFLUprK4AvVhsdag==",
	"hBlzKg78ajaZuTE0VLzDDg==",
	"9FvVhtFLUs0KnA3Kprsdyg==",
	"d2ViUmVtZW1iZXJNZUtleQ==",
	"yNeUgSzL/CfiWw1GALg6Ag==",
	"NGk/3cQ6F5/UNPRh8LpMIg==",
	"4BvVhmFLUs0KTA3Kprsdag==",
	"MzVeSkYyWTI2OFVLZjRzZg==",
	"empodDEyMwAAAAAAAAAAAA==",
	"A7UzJgh1+EWj5oBFi+mSgw==",
	"YTM0NZomIzI2OTsmIzM0NTueYQ==",
	"c2hpcm9fYmF0aXMzMgAAAA==",
	"i45FVt72K2kLgvFrJtoZRw==",
	"U3BAbW5nQmxhZGUAAAAAAA==",
	"ZnJlc2h6Y24xMjM0NTY3OA==",
	"Jt3C93kMR9D5e8QzwfsiMw==",
	"MTIzNDU2NzgxMjM0NTY3OA==",
	"vXP33AonIp9bFwGl7aT7rA==",
	"V2hhdCBUaGUgSGVsbAAAAA==",
	"Z3h6eWd4enklMjElMjElMjE=",
	"Q01TX0JGTFlLRVlfMjAxOQ==",
	"ZAvph3dsQs0FSL3SDFAdag==",
	"Is9zJ3pzNh2cgTHB4ua3+Q==",
	"NsZXjXVklWPZwOfkvk6kUA==",
	"GAevYnznvgNCURavBhCr1w==",
	"66v1O8keKNV3TTcGPK1wzg==",
	"SDKOLKn2J1j/2BHjeZwAoQ==",
}

var randKey []byte

func init() {
	rand.NewSource(time.Now().UnixNano())
	randKey, _ = codec.DecodeBase64(keyList[rand.Intn(len(keyList))])
}

func (s *VulinServer) registerMockVulShiro() {
	var router = s.router

	router.HandleFunc("/shiro/cbc", func(writer http.ResponseWriter, request *http.Request) {
		failNow := func(writer http.ResponseWriter, request *http.Request) {
			cookie := http.Cookie{
				Name:     "rememberMe",
				Value:    "deleteMe",                         // 设置 cookie 的值
				Expires:  time.Now().Add(7 * 24 * time.Hour), // 设置过期时间
				HttpOnly: false,                              // 仅限 HTTP 访问，不允许 JavaScript 访问
			}
			http.SetCookie(writer, &cookie)
			writer.WriteHeader(200)
			return
		}
		successNow := func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			return
		}
		rememberMe, err := request.Cookie("rememberMe")
		if err != nil { // 请求没有cookie 那就设置一个
			failNow(writer, request)
			return
		}
		cookieVal, _ := codec.DecodeBase64(rememberMe.Value)
		if len(cookieVal) > len(randKey) {
			cookieVal = cookieVal[len(randKey):]
		} else { // 第一次探测请求
			failNow(writer, request)
			return
		}

		payload, err := codec.AESCBCDecrypt(randKey, cookieVal, nil)
		if err != nil || payload == nil { // key不对返回deleteMe
			failNow(writer, request)
			return
		}
		payload, err = codec.MustPKCS5UnPadding(payload)
		if err != nil || payload == nil { // key不对返回deleteMe
			failNow(writer, request)
			return
		}
		javaObject, err := yso.GetJavaObjectFromBytes(payload)
		if err != nil { // 反序列化出错返回
			failNow(writer, request)
			return
		}
		if strings.Contains(string(javaObject.Marshal()), "org.apache.shiro.subject.SimplePrincipalCollection") {
			successNow(writer, request)
		} else {
			failNow(writer, request)
			return
		}
		writer.WriteHeader(200)
		return
	})
	router.HandleFunc("/shiro/gcm", func(writer http.ResponseWriter, request *http.Request) {
		failNow := func(writer http.ResponseWriter, request *http.Request) {
			cookie := http.Cookie{
				Name:     "rememberMe",
				Value:    "deleteMe",                         // 设置 cookie 的值
				Expires:  time.Now().Add(7 * 24 * time.Hour), // 设置过期时间
				HttpOnly: false,                              // 仅限 HTTP 访问，不允许 JavaScript 访问
			}
			http.SetCookie(writer, &cookie)
			writer.WriteHeader(200)
			return
		}
		successNow := func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			return
		}
		rememberMe, err := request.Cookie("rememberMe")
		if err != nil { // 请求没有cookie 那就设置一个
			failNow(writer, request)
			return
		}
		cookieVal, _ := codec.DecodeBase64(rememberMe.Value)

		payload, err := codec.AESGCMDecrypt(randKey, cookieVal, nil)
		if err != nil || payload == nil { // key不对返回deleteMe
			failNow(writer, request)
			return
		}
		payload, err = codec.MustPKCS5UnPadding(payload)
		if err != nil || payload == nil { // key不对返回deleteMe
			failNow(writer, request)
			return
		}
		javaObject, err := yso.GetJavaObjectFromBytes(payload)
		if err != nil { // 反序列化出错返回
			failNow(writer, request)
			return
		}
		if strings.Contains(string(javaObject.Marshal()), "org.apache.shiro.subject.SimplePrincipalCollection") {
			successNow(writer, request)
		} else {
			failNow(writer, request)
			return
		}
		writer.WriteHeader(200)
		return
	})
}
