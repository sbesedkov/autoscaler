/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"io"
	klog "k8s.io/klog/v2"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ec2MetaDataServiceUrl          = "http://169.254.169.254"
	ec2PricingServiceUrlTemplate   = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/%s/index.csv"
	ec2PricingServiceUrlTemplateCN = "https://pricing.cn-north-1.amazonaws.com.cn/offers/v1.0/cn/AmazonEC2/current/%s/index.csv"
	staticListLastUpdateTime       = "2020-12-07"
	ec2Arm64Processors             = []string{"AWS Graviton Processor", "AWS Graviton2 Processor"}
)

type attrIndex struct {
  InstanceType  int64 `hdr:"Instance Type"`
  VCPU          int64 `hdr:"vCPU"`
  Memory        int64 `hdr:"Memory"`
  Architecture  int64 `hdr:"Processor Architecture"`
  GPU           int64 `hdr:"GPU"`
}

// GenerateEC2InstanceTypes returns a map of ec2 resources
func GenerateEC2InstanceTypes(region string) (map[string]*InstanceType, error) {
	var pricingUrlTemplate string
	if strings.HasPrefix(region, "cn-") {
		pricingUrlTemplate = ec2PricingServiceUrlTemplateCN
	} else {
		pricingUrlTemplate = ec2PricingServiceUrlTemplate
	}

	instanceTypes := make(map[string]*InstanceType)

	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	for _, p := range partitions {
		for _, r := range p.Regions() {
			if region != "" && region != r.ID() {
				continue
			}

			url := fmt.Sprintf(pricingUrlTemplate, r.ID())
			klog.V(1).Infof("fetching %s\n", url)
			res, err := http.Get(url)
			if err != nil {
				klog.Warningf("Error fetching %s skipping...\n%s\n", url, err)
				continue
			}

			defer res.Body.Close()

            body := csv.NewReader(res.Body)
            body.FieldsPerRecord = -1;

            var header []string
            // Skip first 5 lines and read header
            for i := 0; i < 6; i++ {
                row, err := body.Read()
                if err != nil {
                    header = nil
                    break
                }
                header = row
            }

			if header == nil || len(header) == 0 {
				klog.Warningf("Error parsing %s skipping...\n", url)
				continue
			}

			idx, err := parseHeader(header)
			if err != nil {
			    klog.Warningf("Error parsing header from %s. %s. Skipping\n", url, err)
			    continue
			}

			for {
				attr, err := body.Read()

                if err == io.EOF {
                  break
                }
                if err != nil {
                  klog.Warningf("Error reading %s skipping...\n", url)
                  break
                }
                if (len(attr) != len(header)) {
                  klog.Warningf("Error parsing %s: invalid number of attributes. Skipping...\n", url)
                  break
                }

				if attr[idx.InstanceType] != "" {
				    instanceType := attr[idx.InstanceType]
					instanceTypes[instanceType] = &InstanceType{
						InstanceType: instanceType,
					}
					if attr[idx.Memory] != "" && attr[idx.Memory] != "NA" {
						instanceTypes[instanceType].MemoryMb = parseMemory(attr[idx.Memory])
					}
					if attr[idx.VCPU] != "" {
						instanceTypes[instanceType].VCPU = parseCPU(attr[idx.VCPU])
					}
					if attr[idx.GPU] != "" {
						instanceTypes[instanceType].GPU = parseCPU(attr[idx.GPU])
					}
					if attr[idx.Architecture] != "" {
						instanceTypes[instanceType].Architecture = parseArchitecture(attr[idx.Architecture])
					}
				}
			}
		}
	}

	if len(instanceTypes) == 0 {
		return nil, errors.New("unable to load EC2 Instance Type list")
	}

	return instanceTypes, nil
}

// GetStaticEC2InstanceTypes return pregenerated ec2 instance type list
func GetStaticEC2InstanceTypes() (map[string]*InstanceType, string) {
	return InstanceTypes, staticListLastUpdateTime
}

func parseHeader(header []string) (*attrIndex, error) {
  idx := new(attrIndex)

  headerIndex := make(map[string]int64)
  for i, attr := range header {
    headerIndex[attr] = int64(i)
  }

  rIdx = reflect.Indirect(reflect.ValueOf(idx))

  for i := 0; i < rIdx.NumField(); i++ {
    title := rIdx.Type().Field(i).Tag.Get("hdr")
    if _, exists := headerIndex[title]; !exists {
        return nil, fmt.Errorf("No %s found in header", title)
    }

    rIdx.Field(i).SetInt(headerIndex[title])
  }

  return idx, nil
}

func parseMemory(memory string) int64 {
	reg, err := regexp.Compile("[^0-9\\.]+")
	if err != nil {
		klog.Fatal(err)
	}

	parsed := strings.TrimSpace(reg.ReplaceAllString(memory, ""))
	mem, err := strconv.ParseFloat(parsed, 64)
	if err != nil {
		klog.Fatal(err)
	}

	return int64(mem * float64(1024))
}

func parseCPU(cpu string) int64 {
	i, err := strconv.ParseInt(cpu, 10, 64)
	if err != nil {
		klog.Fatal(err)
	}
	return i
}

func parseArchitecture(archName string) string {
	for _, processor := range ec2Arm64Processors {
		if archName == processor {
			return "arm64"
		}
	}
	return "amd64"
}

// GetCurrentAwsRegion return region of current cluster without building awsManager
func GetCurrentAwsRegion() (string, error) {
	region, present := os.LookupEnv("AWS_REGION")

	if !present {
		c := aws.NewConfig().
			WithEndpoint(ec2MetaDataServiceUrl)
		sess, err := session.NewSession()
		if err != nil {
			return "", fmt.Errorf("failed to create session")
		}
		return ec2metadata.New(sess, c).Region()
	}

	return region, nil
}
