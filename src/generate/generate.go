//go:generate goversioninfo

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"
	"yaml"

	"golang.org/x/oauth2"

	"models"
)

const (
	installBatTemplate = `msiexec /passive /norestart /i %~dp0\DiegoWindows.msi ^{{ if .BbsRequireSsl }}
  BBS_CA_FILE=%~dp0\bbs_ca.crt ^
  BBS_CLIENT_CERT_FILE=%~dp0\bbs_client.crt ^
  BBS_CLIENT_KEY_FILE=%~dp0\bbs_client.key ^{{ end }}
  REP_REQUIRE_TLS={{.RepRequireTls}} ^{{if .RepRequireTls}}
  REP_CA_CERT_FILE=%~dp0\rep_ca.crt ^
  REP_SERVER_CERT_FILE=%~dp0\rep_server.crt ^
  REP_SERVER_KEY_FILE=%~dp0\rep_server.key ^{{ end }}
  CONSUL_DOMAIN={{.ConsulDomain}} ^
  CONSUL_IPS={{.ConsulIPs}} ^
  CF_ETCD_CLUSTER=http://{{.EtcdCluster}}:4001 ^
  STACK=windows2012R2 ^
  REDUNDANCY_ZONE={{.Zone}} ^
  LOGGREGATOR_SHARED_SECRET={{.SharedSecret}} ^
  MACHINE_IP={{.MachineIp}}{{ if .SyslogHostIP }} ^
  SYSLOG_HOST_IP={{.SyslogHostIP}} ^
  SYSLOG_PORT={{.SyslogPort}}{{ end }}{{if .ConsulRequireSSL }} ^
  CONSUL_ENCRYPT_FILE=%~dp0\consul_encrypt.key ^
  CONSUL_CA_FILE=%~dp0\consul_ca.crt ^
  CONSUL_AGENT_CERT_FILE=%~dp0\consul_agent.crt ^
  CONSUL_AGENT_KEY_FILE=%~dp0\consul_agent.key{{end}}{{if .MetronPreferTLS }} ^
  METRON_CA_FILE=%~dp0\metron_ca.crt ^
  METRON_AGENT_CERT_FILE=%~dp0\metron_agent.crt ^
  METRON_AGENT_KEY_FILE=%~dp0\metron_agent.key{{end}}

msiexec /passive /norestart /i %~dp0\GardenWindows.msi ^
  MACHINE_IP={{.MachineIp}}{{ if .SyslogHostIP }} ^
  SYSLOG_HOST_IP={{.SyslogHostIP}} ^
  SYSLOG_PORT={{.SyslogPort}}{{ end }}`
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of generate:\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var (
		cfManifest    string
		outputDir     string
		boshServerUrl string
		machineIp     string
	)
	flag.StringVar(&cfManifest, "manifest", "", "Path to CF manifest file")
	flag.StringVar(&outputDir, "outputDir", "", "Directory where the generated install script and certs will be created")
	flag.StringVar(&boshServerUrl, "boshUrl", "", "Bosh director URL e.g. https://admin:admin@bosh.example:25555")
	flag.StringVar(&machineIp, "machineIp", "", "(optional) IP address of this cell")

	flag.Parse()
	if outputDir == "" || (boshServerUrl == "" && cfManifest == "") {
		usage()
	}
	if boshServerUrl != "" && cfManifest != "" {
		fmt.Fprintln(os.Stderr, "Error: only boshServerUrl or cfManifest may be specified")
		usage()
	}

	var manifest models.Manifest
	if cfManifest != "" {
		manifestContents, err := ioutil.ReadFile(cfManifest)
		err = yaml.Unmarshal(manifestContents, &manifest)
		if err != nil {
			Fatal(err)
		}
	} else {

		u, _ := url.Parse(boshServerUrl)

		_, err := os.Stat(outputDir)
		if err != nil {
			if os.IsNotExist(err) {
				os.MkdirAll(outputDir, 0755)
			}
		}

		bosh := NewBosh(*u)
		bosh.Authorize()

		response := bosh.MakeRequest("/deployments")
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(response.Body)
			if err != nil {
				fmt.Printf("Could not read response from BOSH director.")
				os.Exit(1)
			}

			fmt.Fprintf(os.Stderr, "Unexpected BOSH director response: %v, %v", response.StatusCode, buf.String())
			os.Exit(1)
		}

		deployments := []models.IndexDeployment{}
		json.NewDecoder(response.Body).Decode(&deployments)
		idx := GetDiegoDeployment(deployments)
		if idx == -1 {
			fmt.Fprintf(os.Stderr, "BOSH Director does not have exactly one deployment containing a cf and diego release.")
			os.Exit(1)
		}

		response = bosh.MakeRequest("/deployments/" + deployments[idx].Name)
		defer response.Body.Close()

		deployment := models.ShowDeployment{}
		json.NewDecoder(response.Body).Decode(&deployment)

		err = yaml.Unmarshal([]byte(deployment.Manifest), &manifest)
		if err != nil {
			Fatal(err)
		}
	}

	args, err := models.NewInstallerArguments(&manifest)
	if err != nil {
		Fatal(err)
	}

	args.FillEtcdCluster()
	args.FillSharedSecret()
	args.FillMetronAgent()
	args.FillSyslog()
	args.FillConsul()
	args.FillBBS()
	args.FillRep()

	if machineIp == "" {
		consulIp := strings.Split(args.ConsulIPs, ",")[0]
		conn, err := net.Dial("udp", consulIp+":65530")
		Fatal(err)
		machineIp = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	args.FillMachineIp(machineIp)

	generateInstallScript(outputDir, args)
	writeCerts(outputDir, args)
}

func writeCerts(outputDir string, args *models.InstallerArguments) {
	for filename, cert := range args.Certs {
		err := ioutil.WriteFile(path.Join(outputDir, filename), []byte(cert), 0644)
		if err != nil {
			Fatal(err)
		}
	}
}

func Fatal(err interface{}) {
	if err == nil {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
	}
	switch err.(type) {
	case error, string, fmt.Stringer:
		if ok {
			fmt.Fprintf(os.Stderr, "Error (%s:%d): %s\n", file, line, err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
	default:
		if ok {
			fmt.Fprintf(os.Stderr, "Error (%s:%d): %v\n", file, line, err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
	os.Exit(1)
}

func generateInstallScript(outputDir string, args *models.InstallerArguments) {
	content := strings.Replace(installBatTemplate, "\n", "\r\n", -1)
	temp := template.Must(template.New("").Parse(content))
	args.Zone = "windows"
	filename := "install.bat"
	file, err := os.OpenFile(path.Join(outputDir, filename), os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	Fatal(err)
	defer file.Close()

	err = temp.Execute(file, args)
	Fatal(err)
}

func GetDiegoDeployment(deployments []models.IndexDeployment) int {
	deploymentIndex := -1

	for i, deployment := range deployments {
		releases := map[string]bool{}
		for _, rel := range deployment.Releases {
			releases[rel.Name] = true
		}

		if releases["cf"] && releases["diego"] && releases["garden-runc"] {
			if deploymentIndex != -1 {
				return -1
			}

			deploymentIndex = i
		}

	}

	return deploymentIndex
}

func NewBosh(endpoint url.URL) *Bosh {
	return &Bosh{
		endpoint: endpoint,
	}
}

type Bosh struct {
	endpoint  url.URL
	authToken string
	authType  string
}

type BoshInfo struct {
	UserAuthentication struct {
		Type    string `json:"type"`
		Options struct {
			Url string `json:"url"`
		} `json:"options"`
	} `json:"user_authentication"`
}

func (b *Bosh) Authorize() {
	if b.endpoint.User == nil {
		log.Fatalln("Director username and password are required.")
	}
	password, _ := b.endpoint.User.Password()
	if password == "" {
		log.Fatalln("Director password is required.")
	}
	resp := b.MakeRequest("/info")
	defer resp.Body.Close()
	var info BoshInfo
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &info)
	b.authType = info.UserAuthentication.Type
	if b.authType == "uaa" {
		tokenEndpoint, err := url.Parse("oauth/token")
		if err != nil {
			log.Fatal(err)
		}
		authEndpoint, err := url.Parse("oauth/authorize")
		if err != nil {
			log.Fatal(err)
		}
		uaaUrl, err := url.Parse(info.UserAuthentication.Options.Url)
		if err != nil {
			log.Fatal(err)
		}
		authURL := uaaUrl.ResolveReference(authEndpoint).String()
		tokenURL := uaaUrl.ResolveReference(tokenEndpoint).String()
		conf := &oauth2.Config{
			ClientID:     "bosh_cli",
			ClientSecret: "",
			Scopes:       []string{"bosh.read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		}

		token, err := conf.PasswordCredentialsToken(nil, b.endpoint.User.Username(), password)
		if err != nil {
			log.Fatal(err)
		}

		b.authToken = token.AccessToken
		b.endpoint.User = nil
	}
}

func (b *Bosh) MakeRequest(path string) *http.Response {
	request, err := http.NewRequest("GET", b.endpoint.String()+path, nil)
	if err != nil {
		log.Fatal(err)
	}
	if b.authType == "uaa" {
		request.Header.Set("Authorization", fmt.Sprintf("bearer %s", b.authToken))
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	http.DefaultClient.Timeout = 10 * time.Second
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalln("Unable to establish connection to BOSH Director.", err)
	}
	return response
}
