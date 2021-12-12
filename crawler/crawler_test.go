package crawler

import "testing"

func TestImageCheck(t *testing.T) {
	//url := "https://gn-web-assets.api.bbc.com/wwhp/20210923-1449-37491ec2b6e5b4c43bda3673e521e8164a789b87/responsive/img/apple-touch/apple-touch-180.jpg"

	url := "https://alerts.ndtv.com/images/web.png"
	t.Log("url")

	if !checkImage(url) {
		t.Errorf("says not a image")
	}
}
