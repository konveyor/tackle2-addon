package command

import (
	"errors"
	"github.com/onsi/gomega"
	"regexp"
	"testing"
)

func TestOptions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	filter := Mask{AuthMask}
	options := Options{}
	options.Add("http://almer:fudd@redhat.com")
	s := options.String(filter)
	g.Expect("http://###:###@redhat.com").To(gomega.Equal(s))
}

func TestFilteringInCommand(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	cmd := Command{
		Path:   "echo",
		Silent: true,
	}
	cmd.Mask = Mask{AuthMask}
	cmd.Options.Add("SEND")
	cmd.Options.Add("http://almer:fudd@redhat.com")
	err := cmd.Run()
	g.Expect(err).To(gomega.BeNil())
	g.Expect("SEND http://###:###@redhat.com\n").To(gomega.Equal(cmd.Output))
}

func TestErrorMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	output := "fatal: Authentication failed for 'https://github.com/konveyor/tackle-testapp.git/"
	mp := ErrorMap{
		ErrorPattern{
			Regex: regexp.MustCompile(`Authentication failed`),
			Error: func(s string) (e error) {
				e = errors.New(s)
				return
			},
		},
	}
	err := mp.Error(nil, "Hello")
	g.Expect(err).To(gomega.BeNil())
	err = mp.Error(nil, output)
	g.Expect(err).ToNot(gomega.BeNil())
}
