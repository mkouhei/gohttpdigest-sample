/*
original is https://github.com/mattbaird/http-digest-auth-client
*/
package main

import (
	"fmt"
	//"io/ioutil"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type myjar struct {
	jar map[string][]*http.Cookie
}

type digestHeaders struct {
	Realm     string
	Qop       string
	Nonce     string
	Opaque    string
	Algorithm string
	HA1       string
	HA2       string
	Cnonce    string
	Path      string
	Nc        int
	Username  string
	Password  string
}

func (p *myjar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	p.jar[u.Host] = cookies
}

func (p *myjar) Cookies(u *url.URL) []*http.Cookie {
	return p.jar[u.Host]
}

func (d *digestHeaders) digestChecksum() {
	switch d.Algorithm {
	case "MD5":
		// A1
		h := md5.New()
		A1 := fmt.Sprintf("%s:%s:%s", d.Username, d.Realm, d.Password)
		io.WriteString(h, A1)
		d.HA1 = fmt.Sprintf("%x", h.Sum(nil))

		// A2
		h = md5.New()
		A2 := fmt.Sprintf("GET:%s", d.Path)
		io.WriteString(h, A2)
		d.HA2 = fmt.Sprintf("%x", h.Sum(nil))
	case "MD5-sess":
	default:
		//token
	}
}

func (d *digestHeaders) Get(uri string) (*http.Response, error) {
	u, _ := url.Parse(uri)
	d.Path = u.Path
	d.digestChecksum()
	response := H(strings.Join([]string{d.HA1, d.Nonce, fmt.Sprintf("%08d", 1),
		d.Cnonce, d.Qop, d.HA2}, ":"))
	AuthHeader := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc=%08d, qop=%s, response="%s", algorithm=%s`,
		d.Username, d.Realm, d.Nonce, d.Path, d.Cnonce, d.Nc, d.Qop, response, d.Algorithm)
	if d.Opaque != "" {
		AuthHeader = fmt.Sprintf(`%s, opaque="%s"`, AuthHeader, d.Opaque)
	}
	log.Println(AuthHeader)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", AuthHeader)
	client := &http.Client{}
	//resp, err := client.Do(req)
	return client.Do(req)
}

func (d *digestHeaders) Auth(username string, password string, uri string) (bool, error, *digestHeaders) {

	client := &http.Client{}
	jar := &myjar{}
	jar.jar = make(map[string][]*http.Cookie)
	client.Jar = jar

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 401 {

		authn := DigestAuthParams(resp)
		d := &digestHeaders{}
		u, _ := url.Parse(uri)
		d.Path = u.Path
		d.Realm = authn["realm"]
		d.Qop = authn["qop"]
		d.Nonce = authn["nonce"]
		d.Opaque = authn["opaque"]
		d.Algorithm = authn["algorithm"]
		d.Cnonce = RandomKey()
		d.Nc = 1
		d.Username = username
		d.Password = password

		resp, err = d.Get(uri)
		if err != nil {
			log.Fatal(err)
		}
		return resp.StatusCode == 200, err, d
		/*
			// HA1 and HA2
			d.digestChecksum()

			// response
			response := H(strings.Join([]string{d.HA1, d.Nonce, fmt.Sprintf("%08d", 1), d.Cnonce, d.Qop, d.HA2}, ":"))

			// now make header
			AuthHeader := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc=%08d, qop=%s, response="%s", algorithm=%s`,
				d.Username, d.Realm, d.Nonce, d.Path, d.Cnonce, d.Nc, d.Qop, response, d.Algorithm)
			if d.Opaque != "" {
				AuthHeader = fmt.Sprintf(`%s, opaque="%s"`, AuthHeader, d.Opaque)
			}

			req.Header.Set("Authorization", AuthHeader)
			resp, err = client.Do(req)
*/
	} else {
		return false, fmt.Errorf("response status code should have been 401, it was %v", resp.StatusCode), nil
	}
}

/*
 Parse Authorization header from the http.Request. Returns a map of
 auth parameters or nil if the header is not a valid parsable Digest
 auth header.
*/
func DigestAuthParams(r *http.Response) map[string]string {
	s := strings.SplitN(r.Header.Get("Www-Authenticate"), " ", 2)
	if len(s) != 2 || s[0] != "Digest" {
		return nil
	}

	result := map[string]string{}
	for _, kv := range strings.Split(s[1], ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.Trim(parts[0], "\" ")] = strings.Trim(parts[1], "\" ")
	}
	return result
}

func RandomKey() string {
	k := make([]byte, 12)
	for bytes := 0; bytes < len(k); {
		n, err := rand.Read(k[bytes:])
		if err != nil {
			panic("rand.Read() failed")
		}
		bytes += n
	}
	return base64.StdEncoding.EncodeToString(k)
}

/*
 H function for MD5 algorithm (returns a lower-case hex MD5 digest)
*/
func H(data string) string {
	digest := md5.New()
	digest.Write([]byte(data))
	return fmt.Sprintf("%x", digest.Sum(nil))
}
