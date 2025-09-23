package goHttpDigestClient

import "testing"
import "github.com/stretchr/testify/assert"
import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	testWwwAuthStr          = `Digest realm="Users", nonce="EIQrqdZGXLGKROqDCs4YoRDtnXzZTthi", qop="auth"`
	testDigestAuthServerURL = "https://discover-contrarious-bluetit.cfapps.io/"
	testServerUsername      = "username"
	testServerPassword      = "password"
)

func TestGetChanllengeFromHeader(t *testing.T) {
	h := http.Header{}
	h.Set("WWW-Authenticate", `Digest realm="Users", nonce="EIQrqdZGXLGKROqDCs4YoRDtnXzZTthi", qop="auth"`)
	authOption := GetChallengeFromHeader(&h)
	assert.Equal(t, "Digest", authOption.GetChallengeItemPure(KEY_AUTH_SCHEMA), "auth type")
	assert.Equal(t, "Users", authOption.GetChallengeItemPure(KEY_REALM), "auth realm")
	assert.Equal(t, "EIQrqdZGXLGKROqDCs4YoRDtnXzZTthi", authOption.GetChallengeItemPure(KEY_NONCE), "auth server nonce")
	assert.Equal(t, "auth", authOption.GetChallengeItemPure(KEY_QOP), "auth qop")
	t.Log("GetAuthInfoFromHeader can get all info from header['www-authenticate']")
}

func TestIsDigestAuth(t *testing.T) {
	h := http.Header{}
	h.Set("WWW-Authenticate", `Digest realm="Users", nonce="EIQrqdZGXLGKROqDCs4YoRDtnXzZTthi", qop="auth"`)
	authOption := GetChallengeFromHeader(&h)
	assert.True(t, authOption.IsDigestAuth(), "auth type is Digest")
	t.Log("IsDigestAuth pass")
}

func TestNewChallenge(t *testing.T) {
	dai := NewChallenge(`Digest realm="Users", nonce="EIQrqdZGXLGKROqDCs4YoRDtnXzZTthi", qop="auth"`)
	assert.Equal(t, "Digest", dai.GetChallengeItemPure(KEY_AUTH_SCHEMA), "auth type")
	assert.Equal(t, "Users", dai.GetChallengeItemPure(KEY_REALM), "auth realm")
	assert.Equal(t, "EIQrqdZGXLGKROqDCs4YoRDtnXzZTthi", dai.GetChallengeItemPure(KEY_NONCE), "auth server nonce")
	assert.Equal(t, "auth", dai.GetChallengeItemPure(KEY_QOP), "auth qop")
	t.Log("create new DigestAuthInfo pass")
}

func TestToWwwHeaderStr(t *testing.T) {
	dai := NewChallenge(testWwwAuthStr)
	t.Log(fmt.Sprintf("target:%s", testWwwAuthStr))
	t.Log(fmt.Sprintf("result:%s", dai.ToAuthorizationStr()))
	t.Log("check output")
}

func TestToMd5(t *testing.T) {
	source := "u:Users:p"
	target := "e098ce4fd4536daadc81bc661181cb19"
	assert.Equal(t, target, toMd5(source), "md5 check")
	t.Log("md5 hash works fine")
}

func TestComputeHA1(t *testing.T) {
	ha1 := computeHa1("u", "Users", "p")
	assert.Equal(t, ha1, "e098ce4fd4536daadc81bc661181cb19", "ha1 should correct")
	t.Log("compute ha1 works fine")
}

func TestComputeHA2(t *testing.T) {
	ha2 := computeHa2("auth", "GET", "/discoverer/clients", "")
	assert.Equal(t, ha2, "b6a6ab4b6c1f11595e0b6c3749bb5ca5", "ha2 should correct")
	t.Log("compute ha2 works fine")
}

func TestComputeResponse(t *testing.T) {
	response, _, _ := computeResponse("auth", "Users", "123", "0000001", "123", "GET", "/discoverer/clients", "", "u", "p")
	assert.Equal(t, response, "e233ecb7ff6ab65ee08bad22b60a3347", "response should equal")
	t.Log("compute response works fine")
}

func TestAuthorize(t *testing.T) {
	req, err := http.NewRequest("GET", testDigestAuthServerURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	opt := &ClientOption{username: testServerUsername, password: testServerPassword}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	www_auth := NewChallenge(res.Header.Get("WWW-Authenticate"))
	www_auth.ComputeResponse(req.Method, req.URL.RequestURI(), "", opt.username, opt.password)
	authorization := www_auth.ToAuthorizationStr()
	req.Header.Set(KEY_AUTHORIZATION, authorization)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "" {
		t.Log("manual test well")
	}
}

func TestClientAuthorize(t *testing.T) {
	req, err := http.NewRequest("GET", testDigestAuthServerURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	opt := &ClientOption{username: testServerUsername, password: testServerPassword}
	res, err := DefaultClient.Do(req, opt)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("status code: %d", res.StatusCode))
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "" {
		t.Log("client test well")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient(testServerUsername,testServerPassword)
	assert.Equal(t,testServerUsername,client.option.username,"option should equal to set");
	assert.Equal(t,testServerPassword,client.option.password,"option should equal to set");
}