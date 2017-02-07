package smoke_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"github.com/cloudfoundry/bosh-utils/system"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"encoding/json"
	"text/template"
)

func TestSmoke(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smoke Suite")
}

var _ = BeforeSuite(func() {
	assertRequiredParams()
	login()
	uploadRelease()
	uploadStemcell()

	switch iaas := os.Getenv("IAAS"); iaas {
	case "vbox":
		updateVboxCloudConfig()
	default:
		updateVsphereCloudConfig()
	}

	deploy()
})

var _ = AfterSuite(func() {
	stdOut, stdErr, exitStatus, err := system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(os.Getenv("BOSH_BINARY_PATH"), "-n",  "-d", "bosh-stemcell-smoke-tests", "delete-deployment")
	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))

	stdOut, stdErr, exitStatus, err = system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(os.Getenv("BOSH_BINARY_PATH"), "-n", "clean-up", "--all")
	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))
})



type VsphereEnvironmentResource struct {
	DNS         string `json:"DNS"`
	Description string `json:"_description"`
	DirectorIP  string `json:"directorIP"`
	Network1 struct {
		VCenterVLAN    string `json:"vCenterVLAN"`
		VCenterCIDR    string `json:"vCenterCIDR"`
		VCenterGateway string `json:"vCenterGateway"`
		StaticIP1      string `json:"staticIP-1"`
		StaticIP2      string `json:"staticIP-2"`
		ReservedRange  string `json:"reservedRange"`
		StaticRange    string `json:"staticRange"`
		DynamicRange   string `json:"_dynamicRange"`
		VCenterNetmask string `json:"vCenterNetmask"`
	} `json:"network1"`
	Network2 struct {
		VCenterVLAN    string `json:"vCenterVLAN"`
		VCenterCIDR    string `json:"vCenterCIDR"`
		VCenterGateway string `json:"vCenterGateway"`
		StaticIP1      string `json:"staticIP-1"`
		ReservedRange  string `json:"reservedRange"`
		StaticRange    string `json:"staticRange"`
		DynamicRange   string `json:"_dynamicRange"`
	} `json:"network2"`
	BoshVsphereVcenterDc      string
	BoshVsphereVcenterCluster string
}

func login() {
	stdOut, stdErr, exitStatus, err := system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(os.Getenv("BOSH_BINARY_PATH"), "login")
	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))
}

func deploy() {
	syslogReleaseVersion, err := ioutil.ReadFile("../syslog-release/version")
	Expect(err).ToNot(HaveOccurred())
	stemcellVersion, err := ioutil.ReadFile("../stemcell/version")
	Expect(err).ToNot(HaveOccurred())
	tempFile, err := ioutil.TempFile(os.TempDir(), "manifest")
	Expect(err).ToNot(HaveOccurred())
	contents, err := ioutil.ReadFile("../assets/manifest.yml")
	Expect(err).ToNot(HaveOccurred())

	template, err := template.New("syslog-release").Parse(string(contents))
	Expect(err).ToNot(HaveOccurred())
	err = template.Execute(tempFile, struct {
		SyslogReleaseVersion string
		StemcellVersion      string
	}{
		SyslogReleaseVersion: string(syslogReleaseVersion),
		StemcellVersion:      string(stemcellVersion),
	})
	Expect(err).ToNot(HaveOccurred())
	manifestPath, err := filepath.Abs(tempFile.Name())
	Expect(err).ToNot(HaveOccurred())
	stdOut, stdErr, exitStatus, err := system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(os.Getenv("BOSH_BINARY_PATH"), "-n", "-d", "bosh-stemcell-smoke-tests", "deploy", manifestPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))
}

func updateVboxCloudConfig() {
	stdOut, stdErr, exitStatus, err := system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(
		os.Getenv("BOSH_BINARY_PATH"),
		"-n",
		"update-cloud-config",
		"../assets/vbox/cloud-config.yml",
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))
}

func updateVsphereCloudConfig() {
	environmentResource := &VsphereEnvironmentResource{}
	metadataContents, err := ioutil.ReadFile("../environment/metadata")

	err = json.Unmarshal(metadataContents, environmentResource)
	Expect(err).ToNot(HaveOccurred())
	environmentResource.BoshVsphereVcenterDc = os.Getenv("BOSH_VSPHERE_VCENTER_DC")
	environmentResource.BoshVsphereVcenterCluster = os.Getenv("BOSH_VSPHERE_VCENTER_CLUSTER")

	stdOut, stdErr, exitStatus, err := system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(
		os.Getenv("BOSH_BINARY_PATH"),
		"-n",
		"update-cloud-config",
		"../assets/vsphere/cloud-config.yml",
		"-v", "vcenter_dc="+environmentResource.BoshVsphereVcenterDc,
		"-v", "vcenter_cluster="+environmentResource.BoshVsphereVcenterCluster,
		"-v", "internal_cidr="+environmentResource.Network1.VCenterCIDR,
		"-v", "internal_reserved=["+environmentResource.Network1.ReservedRange+"]",
		"-v", "internal_gw="+environmentResource.Network1.VCenterGateway,
		"-v", "internal_vcenter_vlan="+environmentResource.Network1.VCenterVLAN,
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))
}

func uploadRelease() {
	stdOut, stdErr, exitStatus, err := system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(os.Getenv("BOSH_BINARY_PATH"), "upload-release", os.Getenv("SYSLOG_RELEASE_PATH"))
	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))
}

func assertRequiredParams() {
	_, ok := os.LookupEnv("BOSH_ENVIRONMENT")
	Expect(ok).To(BeTrue(), "BOSH_ENVIRONMENT was not set")
	_, ok = os.LookupEnv("BOSH_CLIENT")
	Expect(ok).To(BeTrue(), "BOSH_CLIENT was not set")
	_, ok = os.LookupEnv("BOSH_CLIENT_SECRET")
	Expect(ok).To(BeTrue(), "BOSH_CLIENT_SECRET was not set")
	_, ok = os.LookupEnv("BOSH_VSPHERE_VCENTER_DC")
	Expect(ok).To(BeTrue(), "BOSH_VSPHERE_VCENTER_DC was not set")
	_, ok = os.LookupEnv("BOSH_VSPHERE_VCENTER_CLUSTER")
	Expect(ok).To(BeTrue(), "BOSH_VSPHERE_VCENTER_CLUSTER was not set")
	_, ok = os.LookupEnv("BOSH_BINARY_PATH")
	Expect(ok).To(BeTrue(), "BOSH_BINARY_PATH was not set")
	_, ok = os.LookupEnv("SYSLOG_RELEASE_PATH")
	Expect(ok).To(BeTrue(), "SYSLOG_RELEASE_PATH was not set")
	_, ok = os.LookupEnv("STEMCELL_PATH")
	Expect(ok).To(BeTrue(), "STEMCELL_PATH was not set")
}

func uploadStemcell() {
	stdOut, stdErr, exitStatus, err := system.NewExecCmdRunner(boshlog.NewLogger(boshlog.LevelDebug)).RunCommand(os.Getenv("BOSH_BINARY_PATH"), "upload-stemcell", os.Getenv("STEMCELL_PATH"))
	Expect(err).ToNot(HaveOccurred())
	Expect(exitStatus).To(Equal(0), fmt.Sprintf("stdOut: %s \n stdErr: %s", stdOut, stdErr))
}