package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Happy Path Installation Tests", func() {
	BeforeEach(func() {
		os.Chdir(kisPath)
	})
	Describe("Calling installer with no input", func() {
		It("should output help text", func() {
			c := exec.Command("./kismatic")
			helpbytes, helperr := c.Output()
			Expect(helperr).To(BeNil())
			helpText := string(helpbytes)
			Expect(helpText).To(ContainSubstring("Usage"))
		})
	})

	Describe("Calling installer with 'install plan'", func() {
		Context("and just hitting enter", func() {
			It("should result in the output of a well formed default plan file", func() {
				By("Outputing a file")
				c := exec.Command("./kismatic", "install", "plan")
				helpbytes, helperr := c.Output()
				Expect(helperr).To(BeNil())
				helpText := string(helpbytes)
				Expect(helpText).To(ContainSubstring("Generating installation plan file template"))
				Expect(helpText).To(ContainSubstring("3 etcd nodes"))
				Expect(helpText).To(ContainSubstring("2 master nodes"))
				Expect(helpText).To(ContainSubstring("3 worker nodes"))
				Expect(helpText).To(ContainSubstring("2 ingress nodes"))
				Expect(helpText).To(ContainSubstring("0 storage nodes"))

				Expect(FileExists("kismatic-cluster.yaml")).To(Equal(true))

				By("Outputing a file with valid YAML")
				yamlBytes, err := ioutil.ReadFile("kismatic-cluster.yaml")
				if err != nil {
					Fail("Could not read cluster file")
				}
				yamlBlob := string(yamlBytes)

				planFromYaml := ClusterPlan{}

				unmarshallErr := yaml.Unmarshal([]byte(yamlBlob), &planFromYaml)
				if unmarshallErr != nil {
					Fail("Could not unmarshall cluster yaml: %v")
				}
			})
		})
	})

	Describe("Calling installer with a plan targeting bad infrastructure", func() {
		Context("Using a 1/1/1 Ubuntu 16.04 layout pointing to bad ip addresses", func() {
			It("should bomb validate and apply", func() {
				if !completesInTime(installKismaticWithABadNode, 600*time.Second) {
					Fail("It shouldn't take 600 seconds for Kismatic to fail with bad nodes.")
				}
			})
		})
	})

	Describe("Installing with package installation enabled", func() {
		installOpts := installOptions{
			allowPackageInstallation: true,
		}
		Context("Targeting AWS infrastructure", func() {
			Context("using a 1/1/1/1/1 layout with Ubuntu 16.04 LTS", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1, 1, 1}, Ubuntu1604LTS, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
						err = verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
						Expect(err).ToNot(HaveOccurred())
						testVolumeAdd(nodes.master[0], sshKey)
					})
				})
			})
			Context("using a 1/1/1/1/1 layout with CentOS 7", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1, 1, 1}, CentOS7, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
						err = verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
						Expect(err).ToNot(HaveOccurred())
						testVolumeAdd(nodes.master[0], sshKey)
					})
				})
			})
			Context("using a 1/1/1/1/1 layout with RedHat 7", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1, 1, 1}, RedHat7, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
						err = verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
						Expect(err).ToNot(HaveOccurred())
						testVolumeAdd(nodes.master[0], sshKey)
					})
				})
			})
			Context("using a 3/2/3/2 layout with CentOS 7", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{3, 2, 3, 2, 0}, CentOS7, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
						err = verifyIngressNodes(nodes.master[0], nodes.ingress, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
			Context("using a 1/2/1 layout with CentOS 7, with DNS", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructureAndDNS(NodeCount{1, 2, 1, 0, 0}, CentOS7, provisioner, func(nodes provisionedNodes, sshKey string) {
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
						err = verifyMasterNodeFailure(nodes, provisioner, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Installing against a minikube layout", func() {
		Context("Targeting AWS infrastructure", func() {
			Context("Using CentOS 7", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithMiniInfrastructure(CentOS7, provisioner, func(node NodeDeets, sshKey string) {
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())
						err = verifyIngressNodes(node, []NodeDeets{node}, sshKey)
						Expect(err).ToNot(HaveOccurred())
						testVolumeAdd(node, sshKey)
					})
				})
			})
			Context("Using Ubuntu 16.04 LTS", func() {
				ItOnAWS("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithMiniInfrastructure(Ubuntu1604LTS, provisioner, func(node NodeDeets, sshKey string) {
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())
						err = verifyIngressNodes(node, []NodeDeets{node}, sshKey)
						Expect(err).ToNot(HaveOccurred())
						testVolumeAdd(node, sshKey)
					})
				})
			})
		})

		Context("Targeting Packet Infrastructure", func() {
			Context("Using CentOS 7", func() {
				ItOnPacket("should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithMiniInfrastructure(CentOS7, provisioner, func(node NodeDeets, sshKey string) {
						err := installKismaticMini(node, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Installing with package installation disabled", func() {
		installOpts := installOptions{
			allowPackageInstallation: false,
		}
		Context("Targeting AWS infrastructure", func() {
			Context("Using a 1/1/1 layout with Ubuntu 16.04 LTS", func() {
				ItOnAWS("Should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1, 0, 0}, Ubuntu1604LTS, provisioner, func(nodes provisionedNodes, sshKey string) {
						By("Installing the Kismatic RPMs")
						InstallKismaticPackages(nodes, Ubuntu1604LTS, sshKey)
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			Context("Using a 1/1/1 CentOS 7 layout", func() {
				ItOnAWS("Should result in a working cluster", func(provisioner infrastructureProvisioner) {
					WithInfrastructure(NodeCount{1, 1, 1, 0, 0}, CentOS7, provisioner, func(nodes provisionedNodes, sshKey string) {
						By("Installing the Kismatic RPMs")
						InstallKismaticPackages(nodes, CentOS7, sshKey)
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Installing with private Docker registry", func() {
		Context("Using a 1/1/1 CentOS 7 layout", func() {
			nodeCount := NodeCount{1, 1, 1, 0, 0}
			distro := CentOS7

			Context("Using the auto-configured docker registry", func() {
				ItOnAWS("should result in a working cluster", func(aws infrastructureProvisioner) {
					WithInfrastructure(nodeCount, distro, aws, func(nodes provisionedNodes, sshKey string) {
						installOpts := installOptions{
							allowPackageInstallation:    true,
							autoConfigureDockerRegistry: true,
						}
						err := installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			Context("Using a custom registry provided by the user", func() {
				ItOnAWS("should result in a working cluster", func(aws infrastructureProvisioner) {
					WithInfrastructure(nodeCount, distro, aws, func(nodes provisionedNodes, sshKey string) {
						By("Installing an external Docker registry on one of the etcd nodes")
						dockerRegistryPort := 8443
						caFile, err := deployDockerRegistry(nodes.etcd[0], dockerRegistryPort, sshKey)
						Expect(err).ToNot(HaveOccurred())
						installOpts := installOptions{
							allowPackageInstallation: true,
							dockerRegistryCAPath:     caFile,
							dockerRegistryIP:         nodes.etcd[0].PrivateIP,
							dockerRegistryPort:       dockerRegistryPort,
						}
						err = installKismatic(nodes, installOpts, sshKey)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Using the hosts file modification option", func() {
		ItOnAWS("should result in a working cluster", func(aws infrastructureProvisioner) {
			WithInfrastructure(NodeCount{1, 1, 2, 0, 0}, CentOS7, aws, func(nodes provisionedNodes, sshKey string) {
				By("Setting the hostnames to be different than the actual ones")
				loadBalancedFQDN := nodes.master[0].PublicIP
				nodes.etcd[0].Hostname = "etcd01"
				nodes.master[0].Hostname = "master01"
				nodes.worker[0].Hostname = "worker01"
				nodes.worker[1].Hostname = "worker02"

				plan := PlanAWS{
					AllowPackageInstallation: true,
					Etcd:                nodes.etcd,
					Master:              nodes.master,
					MasterNodeFQDN:      loadBalancedFQDN,
					MasterNodeShortName: loadBalancedFQDN,
					Worker:              nodes.worker[0:1],
					SSHKeyFile:          sshKey,
					SSHUser:             nodes.master[0].SSHUser,
					ModifyHostsFiles:    true,
				}

				By("Installing kismatic with bogus hostnames that are added to hosts files")
				err := installKismaticWithPlan(plan, sshKey)
				Expect(err).ToNot(HaveOccurred())

				By("Adding a worker with a bogus hostname that is added to hosts files")
				err = addWorkerToKismaticMini(nodes.worker[1])
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Installing on machines with no internet", func() {
		Context("with kismatic packages installed", func() {
			ItOnAWS("should result in a working cluster", func(aws infrastructureProvisioner) {
				WithMiniInfrastructure(CentOS7, aws, func(node NodeDeets, sshKey string) {
					By("Installing the RPMs on the node")
					theNode := []NodeDeets{node}
					nodes := provisionedNodes{
						etcd:    theNode,
						master:  theNode,
						worker:  theNode,
						ingress: theNode,
					}
					InstallKismaticPackages(nodes, CentOS7, sshKey)

					By("Verifying connectivity to google.com")
					err := runViaSSH([]string{"curl --head www.google.com"}, theNode, sshKey, 1*time.Minute)
					FailIfError(err, "Failed to curl google")

					By("Blocking all outbound connections")
					allowPorts := "8888,2379,6666,2380,6660,6443,8443,80,443" // ports needed/checked by inspector
					cmd := []string{
						"sudo iptables -A OUTPUT -o lo -j ACCEPT",                                                         // allow loopback
						"sudo iptables -A OUTPUT -p tcp --sport 22 -m state --state ESTABLISHED -j ACCEPT",                // allow SSH
						fmt.Sprintf("sudo iptables -A OUTPUT -p tcp --match multiport --sports %s -j ACCEPT", allowPorts), // allow inspector
						"sudo iptables -A OUTPUT -s 172.16.0.0/16 -j ACCEPT",
						"sudo iptables -A OUTPUT -d 172.16.0.0/16 -j ACCEPT", // Allow pod network
						"sudo iptables -P OUTPUT DROP",                       // drop everything else
					}
					err = runViaSSH(cmd, theNode, sshKey, 1*time.Minute)
					FailIfError(err, "Failed to create iptable rule")

					By("Verifying that connections are blocked")
					err = runViaSSH([]string{"curl --max-time 5 www.google.com"}, theNode, sshKey, 1*time.Minute)
					if err == nil {
						Fail("was able to ping google with outgoing connections blocked")
					}

					By("Running kismatic")
					installOpts := installOptions{
						allowPackageInstallation:    false,
						modifyHostsFiles:            true,
						autoConfigureDockerRegistry: true,
					}

					err = installKismatic(nodes, installOpts, sshKey)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

	})
})
