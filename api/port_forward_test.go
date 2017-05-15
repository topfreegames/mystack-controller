package api_test

import (
	"encoding/json"
	"net"

	"github.com/Sirupsen/logrus"
	. "github.com/topfreegames/mystack-controller/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PortForward", func() {
	Describe("Proxy", func() {
		It("should copy from one conn to another", func() {
			connA, connB := net.Pipe()
			l := logrus.New()

			Proxy(connA, connB, l)
			connA.Write([]byte("message"))

			bts := make([]byte, 7)
			n, err := connB.Read(bts)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bts)).To(Equal("message"))
			Expect(n).To(Equal(7))
		})
	})

	Describe("Read", func() {
		It("should read object from conn", func() {
			connA, connB := net.Pipe()
			defer connA.Close()
			defer connB.Close()
			l := logrus.New()
			obj := map[string]interface{}{
				"token":   "i-am-a-token",
				"service": "svc1",
			}
			bts, err := json.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			bts = append(bts, '\n')
			go connA.Write(bts)

			token, service, err := Read(connB, l)
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("i-am-a-token"))
			Expect(service).To(Equal("svc1"))
		})
	})
})
