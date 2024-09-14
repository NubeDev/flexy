package subjects

import (
	"fmt"
	"testing"
)

func TestNewSubjectBuilder(t *testing.T) {
	sbProxy := NewSubjectBuilder("<global_uuid>", "<app_id>", IsProxy)
	subjectProxy := sbProxy.BuildSubject("get", "points", "all")
	fmt.Println("Subject with Proxy:", subjectProxy)
	// Output: <global_uuid>.proxy.<app_id>.get.points.all

	// Example when dontUseProxy is true (not using proxy and swapping app_id with global_uuid)
	sbNoProxy := NewSubjectBuilder("abc", "bios", IsProxy)
	subjectNoProxy := sbNoProxy.BuildSubject("get", "points", "all")
	fmt.Println("Subject without Proxy:", subjectNoProxy)
	// Output: <global_uuid>.get.points.all

	// Building a message
	payload := map[string]interface{}{
		"filter": map[string]string{
			"tag": "abc",
		},
	}
	message, _ := BuildMessage(payload)
	fmt.Println("Message:", message)
	subject := "abc.proxy.ros.get.points"
	fmt.Println(GetSubjectParts(subject))

}
