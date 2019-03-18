package spotinst

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spotinst/spotinst-sdk-go/service/mrscaler"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/terraform-providers/terraform-provider-spotinst/spotinst/commons"
	"log"
	"testing"
	"time"
)

var clusterID *string = nil

func createMRScalerAWSResourceName(name string) string {
	return fmt.Sprintf("%v.%v", string(commons.MRScalerAWSResourceName), name)
}

func testMRScalerAWSDestroy(s *terraform.State) error {
	client := testAccProviderAWS.Meta().(*Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != string(commons.MRScalerAWSResourceName) {
			continue
		}
		input := &mrscaler.ReadScalerInput{ScalerID: spotinst.String(rs.Primary.ID)}
		resp, err := client.mrscaler.Read(context.Background(), input)
		if err == nil && resp != nil && resp.Scaler != nil {
			return fmt.Errorf("scaler still exists")
		}
	}
	return nil
}

func testCheckMRScalerAWSAttributes(scaler *mrscaler.Scaler, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if spotinst.StringValue(scaler.Name) != expectedName {
			return fmt.Errorf("bad content: %v", scaler.Name)
		}
		return nil
	}
}

func testCheckMRScalerAWSExists(scaler *mrscaler.Scaler, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}
		client := testAccProviderAWS.Meta().(*Client)
		input := &mrscaler.ReadScalerInput{ScalerID: spotinst.String(rs.Primary.ID)}
		resp, err := client.mrscaler.Read(context.Background(), input)
		if err != nil {
			return err
		}
		if spotinst.StringValue(resp.Scaler.Name) != rs.Primary.Attributes["name"] {
			return fmt.Errorf("mrscaler not found: %+v,\n %+v\n", resp.Scaler, rs.Primary.Attributes)
		}
		*scaler = *resp.Scaler
		return nil
	}
}

type MRScalerAWSConfigMetaData struct {
	variables            string
	provider             string
	scalerName           string
	clusterID            string
	strategy             string
	strategyConfig       string
	cluster              string
	masterGroup          string
	coreGroup            string
	taskGroup            string
	fieldsToAppend       string
	newCluster           bool
	clonedCluster        bool
	wrappedCluster       bool
	updateBaselineFields bool
}

func createMRScalerAWSTerraform(mcm *MRScalerAWSConfigMetaData) string {
	time.Sleep(6 * time.Second)
	if mcm == nil {
		return ""
	}

	if mcm.provider == "" {
		mcm.provider = "aws"
	}

	template :=
		`provider "aws" {
	 token   = "fake"
	 account = "fake"
	}
	`
	format := ""

	if mcm.newCluster {
		mcm.strategy = "new"

		if mcm.updateBaselineFields {
			format = testMRScalerAWSBaseline_Update
		} else {
			format = testMRScalerAWSBaseline_Create
		}

		if mcm.cluster == "" {
			mcm.cluster = testMRScalerAWSCluster_Create
		}

		if mcm.strategyConfig == "" {
			mcm.strategyConfig = testMRScalerAWSStrategy_Create
		}

		if mcm.masterGroup == "" {
			mcm.masterGroup = testMRScalerAWSMasterGroup_Create
		}

		if mcm.coreGroup == "" {
			mcm.coreGroup = testMRScalerAWSCoreGroup_Create
		}

		if mcm.taskGroup == "" {
			mcm.taskGroup = testMRScalerAWSTaskGroup_Create
		}

		template += fmt.Sprintf(format,
			mcm.scalerName,
			mcm.provider,
			mcm.scalerName,
			mcm.strategy,
			mcm.strategyConfig,
			mcm.cluster,
			mcm.masterGroup,
			mcm.coreGroup,
			mcm.taskGroup,
			mcm.fieldsToAppend,
		)
	}

	if mcm.clonedCluster {
		mcm.strategy = "clone"
		mcm.clusterID = "j-TD2G92URMWZX"

		if mcm.updateBaselineFields {
			format = testMRScalerAWSBaselineCloned_Update
		} else {
			format = testMRScalerAWSBaselineCloned_Create
		}

		if mcm.strategyConfig == "" {
			mcm.strategyConfig = testMRScalerAWSStrategy_Create
		}

		if mcm.masterGroup == "" {
			mcm.masterGroup = testMRScalerAWSMasterGroup_Create
		}

		if mcm.coreGroup == "" {
			mcm.coreGroup = testMRScalerAWSCoreGroup_Create
		}

		if mcm.taskGroup == "" {
			mcm.taskGroup = testMRScalerAWSTaskGroup_Create
		}

		template += fmt.Sprintf(format,
			mcm.scalerName,
			mcm.provider,
			mcm.scalerName,
			mcm.strategy,
			mcm.clusterID,
			mcm.strategyConfig,
			mcm.masterGroup,
			mcm.coreGroup,
			mcm.taskGroup,
			mcm.fieldsToAppend,
		)
	}

	log.Printf("Terraform [%v] template:\n%v", mcm.scalerName, template)
	return template
}

// region MRScalerAWS: Baseline
func TestAccSpotinstMRScalerAWSNewCluster_Baseline(t *testing.T) {
	scalerName := "mrscaler-baseline"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					newCluster: true,
					//updateBaselineFields: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
				),
			},
		},
	})
}

const testMRScalerAWSBaseline_Create = `
resource "` + string(commons.MRScalerAWSResourceName) + `" "%v" {
 provider = "%v"

 name               = "%v"
 description        = "test creating a new cluster"
 availability_zones = ["us-west-2b:subnet-1ba25052"]
 strategy           = "%v"
 region             = "us-west-2"

 %v
 %v
 %v
 %v
 %v
 %v
}
`

const testMRScalerAWSBaseline_Update = `
resource "` + string(commons.MRScalerAWSResourceName) + `" "%v" {
 provider = "%v"

 name               = "%v"
 description        = "test updating a created cluster"
 availability_zones = ["us-west-2b:subnet-1ba25052"]
 strategy           = "%v"
 region             = "us-west-2"

 %v
 %v
 %v
 %v
 %v
 %v
}
`

// endregion

// region Strategy

func TestAccSpotinstMRScalerAWSNewCluster_Strategy(t *testing.T) {
	scalerName := "mrscaler-new_cluster-strategy"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					strategyConfig: testMRScalerAWSStrategy_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-5.17.0"),
					//resource.TestCheckResourceAttr(resourceName, "retries", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout", "15"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout_action", "terminate"),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					strategyConfig: testMRScalerAWSStrategy_Update,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-5.17.0"),
					//resource.TestCheckResourceAttr(resourceName, "retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout", "20"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout_action", "terminate"),
				),
			},
		},
	})
}

const testMRScalerAWSStrategy_Create = `
// --- STRATEGY ------------
 release_label = "emr-5.17.0"
 //retries     = 1

 provisioning_timeout = {
   timeout        = 15
   timeout_action = "terminate"
 }
// -------------------------
`

const testMRScalerAWSStrategy_Update = `
// --- STRATEGY ------------
 release_label = "emr-5.17.0"
 //retries     = 3

 provisioning_timeout = {
   timeout        = 20
   timeout_action = "terminate"
 }
// -------------------------
`

// endregion

// region Cluster
func TestAccSpotinstMRScalerAWSNewCluster_Cluster(t *testing.T) {
	scalerName := "mrscaler-new_cluster-cluster"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					cluster:    testMRScalerAWSCluster_Create,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "log_uri", "s3://sorex-job-status"),
					resource.TestCheckResourceAttr(resourceName, "additional_info", "{'test':'more information'}"),
					resource.TestCheckResourceAttr(resourceName, "job_flow_role", "EMR_EC2_DefaultRole"),
					//resource.TestCheckResourceAttr(resourceName, "cluster.0.security_config", "test-config-jeffrey"),
					//resource.TestCheckResourceAttr(resourceName,"cluster.0.service_role",""),
					resource.TestCheckResourceAttr(resourceName, "termination_protected", "false"),
					resource.TestCheckResourceAttr(resourceName, "keep_job_flow_alive", "true"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					cluster:    testMRScalerAWSCluster_Update,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "log_uri", "s3://sorex-job-status"),
					resource.TestCheckResourceAttr(resourceName, "additional_info", "{'test':'more information'}"),
					resource.TestCheckResourceAttr(resourceName, "job_flow_role", "EMR_EC2_DefaultRole"),
					//resource.TestCheckResourceAttr(resourceName, "cluster.0.security_config", "test-config-jeffrey"),
					//resource.TestCheckResourceAttr(resourceName,"cluster.0.service_role",""),
					resource.TestCheckResourceAttr(resourceName, "termination_protected", "false"),
					resource.TestCheckResourceAttr(resourceName, "keep_job_flow_alive", "true"),
				),
			},
		},
	})
}

const testMRScalerAWSCluster_Create = `
 // --- CLUSTER ------------
    log_uri = "s3://sorex-job-status"
    additional_info = "{'test':'more information'}"
    job_flow_role = "EMR_EC2_DefaultRole"
    //security_config = "test-config-jeffrey"
    //service_role = "fake"
    termination_protected = false
    keep_job_flow_alive = true
 // -------------------------
`

const testMRScalerAWSCluster_Update = `
 // --- CLUSTER -------------
    log_uri = "s3://sorex-job-status"
    additional_info = "{'test':'more information'}"
    job_flow_role = "EMR_EC2_DefaultRole"
    //security_config = "test-config-jeffrey"
    //service_role = "fake"
    termination_protected = false
    keep_job_flow_alive = true
 // -------------------------
`

// endregion

// region Instance Groups

func TestAccSpotinstMRScalerAWSNewCluster_MasterGroup(t *testing.T) {
	scalerName := "mrscaler-master-group"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:  scalerName,
					masterGroup: testMRScalerAWSMasterGroup_Create,
					newCluster:  true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "master_instance_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "master_lifecycle", "SPOT"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_optimized", "true"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_block_device.1008334328.volumes_per_instance", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_block_device.1008334328.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_block_device.1008334328.size_in_gb", "30"),
				),
			},
		},
	})
}

const testMRScalerAWSMasterGroup_Create = `
// --- MASTER GROUP -------------
  master_instance_types = ["c3.xlarge"]
  master_lifecycle = "SPOT"
  master_ebs_optimized = true
  master_ebs_block_device = {
    volumes_per_instance = 1
    volume_type = "gp2"
    size_in_gb = 30
    //iops = 1
  }
// ------------------------------
`

// endregion

// region Instance Groups: Task Group

func TestAccSpotinstMRScalerAWSNewCluster_TaskGroup(t *testing.T) {
	scalerName := "mrscaler-task-group"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					taskGroup:  testMRScalerAWSTaskGroup_Create,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_min_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "task_max_size", "30"),
					resource.TestCheckResourceAttr(resourceName, "task_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_lifecycle", "SPOT"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.3329897523.volumes_per_instance", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.3329897523.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.3329897523.size_in_gb", "40"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					taskGroup:  testMRScalerAWSTaskGroup_Update,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_desired_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_lifecycle", "SPOT"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_optimized", "true"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.1008334328.volumes_per_instance", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.1008334328.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.1008334328.size_in_gb", "30"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					taskGroup:  testMRScalerAWSTaskGroup_EmptyFields,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_desired_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_lifecycle", "SPOT"),
				),
			},
		},
	})
}

const testMRScalerAWSTaskGroup_Create = `
// --- TASK GROUP -------------
  task_instance_types = ["c3.xlarge", "c4.xlarge"]
  task_min_size         = 0
  task_max_size         = 30
  task_desired_capacity = 1
  task_lifecycle = "SPOT"
  task_ebs_optimized = false
  task_ebs_block_device = {
    volumes_per_instance = 2
    volume_type = "gp2"
    size_in_gb = 40
    //iops = 1
  }
// ----------------------------
`

const testMRScalerAWSTaskGroup_Update = `
// --- TASK GROUP -------------
  task_instance_types = ["c3.xlarge", "c4.xlarge"]
  task_min_size         = 2
  task_max_size         = 2
  task_desired_capacity = 2
  task_lifecycle = "SPOT"
  task_ebs_optimized = true
  task_ebs_block_device = {
    volumes_per_instance = 1
    volume_type = "gp2"
    size_in_gb = 30
    //iops = 1
  }
// ----------------------------
`

const testMRScalerAWSTaskGroup_EmptyFields = `
// --- TASK GROUP -------------
  task_instance_types = ["c3.xlarge", "c4.xlarge"]
  task_min_size         = 2
  task_max_size         = 2
  task_desired_capacity = 2
  task_lifecycle = "SPOT"
// ----------------------------
`

// endregion

// region Instance Groups: Core Group

func TestAccSpotinstMRScalerAWSNewCluster_CoreGroup(t *testing.T) {
	scalerName := "mrscaler-core-group"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					coreGroup:  testMRScalerAWSCoreGroup_Create,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_lifecycle", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.3329897523.volumes_per_instance", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.3329897523.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.3329897523.size_in_gb", "40"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					coreGroup:  testMRScalerAWSCoreGroup_Update,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_lifecycle", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_optimized", "true"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.1008334328.volumes_per_instance", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.1008334328.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.1008334328.size_in_gb", "30"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName: scalerName,
					coreGroup:  testMRScalerAWSCoreGroup_EmptyFields,
					newCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_lifecycle", "ON_DEMAND"),
				),
			},
		},
	})
}

const testMRScalerAWSCoreGroup_Create = `
// --- CORE GROUP -------------
  core_instance_types = ["c3.xlarge", "c4.xlarge"]
  core_min_size         = 1
  core_max_size         = 1
  core_desired_capacity = 1
  core_lifecycle = "ON_DEMAND"
  core_ebs_optimized = false
  core_ebs_block_device = {
    volumes_per_instance = 2
    volume_type = "gp2"
    size_in_gb = 40
    //iops = 1
  }
// ----------------------------
`

const testMRScalerAWSCoreGroup_Update = `
// --- CORE GROUP -------------
  core_instance_types = ["c3.xlarge", "c4.xlarge"]
  core_min_size         = 1
  core_max_size         = 1
  core_desired_capacity = 1
  core_lifecycle = "ON_DEMAND"
  core_ebs_optimized = true
  core_ebs_block_device = {
    volumes_per_instance = 1
    volume_type = "gp2"
    size_in_gb = 30
    //iops = 1
  }
// ----------------------------
`

const testMRScalerAWSCoreGroup_EmptyFields = `
// --- CORE GROUP -------------
  core_instance_types = ["c3.xlarge", "c4.xlarge"]
  core_min_size         = 1
  core_max_size         = 1
  core_desired_capacity = 1
  core_lifecycle = "ON_DEMAND"
  core_ebs_block_device = {
    volumes_per_instance = 2
    volume_type = "gp2"
    size_in_gb = 30
    //iops = 1
  }
// ----------------------------
`

// endregion

// region Instance Groups: Tags

func TestAccSpotinstMRScalerAWSNewCluster_Tags(t *testing.T) {
	scalerName := "mrscaler-core-group"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testMRScalerAWSTags_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.664003903.key", "Creator"),
					resource.TestCheckResourceAttr(resourceName, "tags.664003903.value", "Terraform"),
				),
			},
		},
	})
}

const testMRScalerAWSTags_Create = `
// --- TAGS -------------
 tags = [
   {
     key = "Creator"
     value = "Terraform"
   }
 ]
// ----------------------
`

// endregion

// region Task Scaling Up Policy

func TestAccSpotinstMRScalerAWSNewCluster_TaskScalingUpPolicies(t *testing.T) {
	scalerName := "mrscaler-task-scaling-up-policy"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t, "aws") },
		Providers:     TestAccProviders,
		CheckDestroy:  testMRScalerAWSDestroy,
		IDRefreshName: resourceName,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testTaskScalingUpPolicy_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.policy_name", "policy-name"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.statistic", "average"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.unit", "percent"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.cooldown", "60"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.dimensions.name", "name-1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.dimensions.value", "value-1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.threshold", "10"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.operator", "gt"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.evaluation_periods", "10"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.period", "60"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.action_type", "adjustment"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.adjustment", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.maximum", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.minimum", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.3662965954.target", ""),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testTaskScalingUpPolicy_Update,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.policy_name", "policy-name-update"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.statistic", "sum"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.unit", "bytes"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.cooldown", "120"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.dimensions.name", "name-1-update"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.dimensions.value", "value-1-update"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.evaluation_periods", "5"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.period", "120"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.action_type", "setMinTarget"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.min_target_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.adjustment", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.maximum", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.minimum", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.90581010.target", ""),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testTaskScalingUpPolicy_EmptyFields,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_up_policy.#", "0"),
				),
			},
		},
	})
}

const testTaskScalingUpPolicy_Create = `
 // --- TASK SCALE UP POLICY -------------
 task_scaling_up_policy = [{
  policy_name = "policy-name"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "average"
  unit = "percent"
  cooldown = 60
  dimensions = {
      name = "name-1"
      value = "value-1"
  }
  threshold = 10

  operator = "gt"
  evaluation_periods = "10"
  period = "60"

  // === MIN TARGET ===================
  //action_type = "setMinTarget"
  //min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  // action_type = "percentageAdjustment"
  action_type = "adjustment"
  adjustment = 1
  // ==================================

  // === UPDATE CAPACITY ==============
  # action_type = "updateCapacity"
  # minimum = 0
  # maximum = 10
  # target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testTaskScalingUpPolicy_Update = `
 // --- TASK SCALE UP POLICY ---------------
 task_scaling_up_policy = [{
  policy_name = "policy-name-update"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "sum"
  unit = "bytes"
  cooldown = 120
  dimensions = {
      name = "name-1-update"
      value = "value-1-update"
  }
  threshold = 5

  operator = "lt"
  evaluation_periods = 5
  period = 120

  // === MIN TARGET ===================
  action_type = "setMinTarget"
  min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  # action_type = "adjustment"
  # action_type = "percentageAdjustment"
  //adjustment = 0
  // ==================================

  // === UPDATE CAPACITY ==============
  # action_type = "updateCapacity"
  # minimum = 0
  # maximum = 10
  # target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testTaskScalingUpPolicy_EmptyFields = `
 // --- TASK SCALE UP POLICY ---------------
 // ----------------------------------------
`

// endregion

// region Task Scaling Down Policy

func TestAccSpotinstMRScalerAWSNewCluster_TaskScalingDownPolicies(t *testing.T) {
	scalerName := "mrscaler-task-scaling-down-policy"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t, "aws") },
		Providers:     TestAccProviders,
		CheckDestroy:  testMRScalerAWSDestroy,
		IDRefreshName: resourceName,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testTaskScalingDownPolicy_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.policy_name", "policy-name"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.statistic", "average"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.unit", "percent"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.cooldown", "60"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.dimensions.name", "name-1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.dimensions.value", "value-1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.threshold", "10"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.evaluation_periods", "10"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.period", "60"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.action_type", "adjustment"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.adjustment", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.maximum", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.minimum", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3776035480.target", ""),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testTaskScalingDownPolicy_Update,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.policy_name", "policy-name-update"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.statistic", "sum"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.unit", "bytes"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.cooldown", "120"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.dimensions.name", "name-1-update"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.dimensions.value", "value-1-update"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.evaluation_periods", "5"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.period", "120"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.action_type", "updateCapacity"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.min_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.adjustment", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.maximum", "10"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.minimum", "0"),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.3217611251.target", "5"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testTaskScalingDownPolicy_EmptyFields,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_scaling_down_policy.#", "0"),
				),
			},
		},
	})
}

const testTaskScalingDownPolicy_Create = `
 // --- TASK SCALE DOWN POLICY -------------
 task_scaling_down_policy = [{
  policy_name = "policy-name"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "average"
  unit = "percent"
  cooldown = 60
  dimensions = {
      name = "name-1"
      value = "value-1"
  }
  threshold = 10

  operator = "lt"
  evaluation_periods = 10
  period = 60

  // === MIN TARGET ===================
  # action_type = "setMinTarget"
  # min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  # action_type = "percentageAdjustment"
  action_type = "adjustment"
  adjustment = 1
  // ==================================

  // === UPDATE CAPACITY ==============
  # action_type = "updateCapacity"
  # minimum = 0
  # maximum = 10
  # target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testTaskScalingDownPolicy_Update = `
 // --- TASK SCALE DOWN POLICY --------------
 task_scaling_down_policy = [{
  policy_name = "policy-name-update"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "sum"
  unit = "bytes"
  cooldown = 120
  dimensions = {
      name = "name-1-update"
      value = "value-1-update"
  }
  threshold = 5

  operator = "lt"
  evaluation_periods = 5
  period = 120

  // === MIN TARGET ===================
  # action_type = "setMinTarget"
  # min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  # action_type = "percentageAdjustment"
  # action_type = "adjustment"
  # adjustment = "MAX(5,10)"
  // ==================================

  // === UPDATE CAPACITY ==============
  action_type = "updateCapacity"
  minimum = 0
  maximum = 10
  target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testTaskScalingDownPolicy_EmptyFields = `
 // --- TASK SCALE DOWN POLICY -------------
 // ----------------------------------------
`

// endregion

// region Core Scaling Up Policy

func TestAccSpotinstMRScalerAWSNewCluster_CoreScalingUpPolicies(t *testing.T) {
	scalerName := "mrscaler-core-scaling-up-policy"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t, "aws") },
		Providers:     TestAccProviders,
		CheckDestroy:  testMRScalerAWSDestroy,
		IDRefreshName: resourceName,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testCoreScalingUpPolicy_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.policy_name", "policy-name"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.statistic", "average"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.unit", "percent"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.cooldown", "60"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.dimensions.name", "name-1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.dimensions.value", "value-1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.threshold", "10"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.operator", "gt"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.evaluation_periods", "10"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.period", "60"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.action_type", "adjustment"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.adjustment", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.maximum", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.minimum", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.3662965954.target", ""),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testCoreScalingUpPolicy_Update,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.policy_name", "policy-name-update"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.statistic", "sum"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.unit", "bytes"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.cooldown", "120"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.dimensions.name", "name-1-update"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.dimensions.value", "value-1-update"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.evaluation_periods", "5"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.period", "120"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.action_type", "setMinTarget"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.min_target_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.adjustment", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.maximum", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.minimum", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.90581010.target", ""),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testCoreScalingUpPolicy_EmptyFields,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_up_policy.#", "0"),
				),
			},
		},
	})
}

const testCoreScalingUpPolicy_Create = `
 // --- CORE SCALE UP POLICY -------------
 core_scaling_up_policy = [{
  policy_name = "policy-name"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "average"
  unit = "percent"
  cooldown = 60
  dimensions = {
      name = "name-1"
      value = "value-1"
  }
  threshold = 10

  operator = "gt"
  evaluation_periods = "10"
  period = "60"

  // === MIN TARGET ===================
  //action_type = "setMinTarget"
  //min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  // action_type = "percentageAdjustment"
  action_type = "adjustment"
  adjustment = 1
  // ==================================

  // === UPDATE CAPACITY ==============
  # action_type = "updateCapacity"
  # minimum = 0
  # maximum = 10
  # target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testCoreScalingUpPolicy_Update = `
 // --- CORE SCALE UP POLICY ---------------
 core_scaling_up_policy = [{
  policy_name = "policy-name-update"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "sum"
  unit = "bytes"
  cooldown = 120
  dimensions = {
      name = "name-1-update"
      value = "value-1-update"
  }
  threshold = 5

  operator = "lt"
  evaluation_periods = 5
  period = 120

  // === MIN TARGET ===================
  action_type = "setMinTarget"
  min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  # action_type = "adjustment"
  # action_type = "percentageAdjustment"
  //adjustment = 0
  // ==================================

  // === UPDATE CAPACITY ==============
  # action_type = "updateCapacity"
  # minimum = 0
  # maximum = 10
  # target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testCoreScalingUpPolicy_EmptyFields = `
 // --- CORE SCALE UP POLICY ---------------
 // ----------------------------------------
`

// endregion

// region Core Scaling Down Policy

func TestAccSpotinstMRScalerAWSNewCluster_CoreScalingDownPolicies(t *testing.T) {
	scalerName := "mrscaler-core-scaling-down-policy"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t, "aws") },
		Providers:     TestAccProviders,
		CheckDestroy:  testMRScalerAWSDestroy,
		IDRefreshName: resourceName,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testCoreScalingDownPolicy_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.policy_name", "policy-name"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.statistic", "average"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.unit", "percent"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.cooldown", "60"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.dimensions.name", "name-1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.dimensions.value", "value-1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.threshold", "10"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.evaluation_periods", "10"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.period", "60"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.action_type", "adjustment"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.adjustment", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.maximum", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.minimum", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3776035480.target", ""),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testCoreScalingDownPolicy_Update,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.policy_name", "policy-name-update"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.statistic", "sum"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.unit", "bytes"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.cooldown", "120"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.dimensions.name", "name-1-update"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.dimensions.value", "value-1-update"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.evaluation_periods", "5"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.period", "120"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.action_type", "updateCapacity"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.min_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.adjustment", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.max_target_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.maximum", "10"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.minimum", "0"),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.3217611251.target", "5"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testCoreScalingDownPolicy_EmptyFields,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_scaling_down_policy.#", "0"),
				),
			},
		},
	})
}

const testCoreScalingDownPolicy_Create = `
 // --- CORE SCALE DOWN POLICY -------------
 core_scaling_down_policy = [{
  policy_name = "policy-name"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "average"
  unit = "percent"
  cooldown = 60
  dimensions = {
      name = "name-1"
      value = "value-1"
  }
  threshold = 10

  operator = "lt"
  evaluation_periods = 10
  period = 60

  // === MIN TARGET ===================
  # action_type = "setMinTarget"
  # min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  # action_type = "percentageAdjustment"
  action_type = "adjustment"
  adjustment = 1
  // ==================================

  // === UPDATE CAPACITY ==============
  # action_type = "updateCapacity"
  # minimum = 0
  # maximum = 10
  # target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testCoreScalingDownPolicy_Update = `
 // --- CORE SCALE DOWN POLICY ---------------
 core_scaling_down_policy = [{
  policy_name = "policy-name-update"
  metric_name = "CPUUtilization"
  namespace = "AWS/EC2"
  statistic = "sum"
  unit = "bytes"
  cooldown = 120
  dimensions = {
      name = "name-1-update"
      value = "value-1-update"
  }
  threshold = 5

  operator = "lt"
  evaluation_periods = 5
  period = 120

  // === MIN TARGET ===================
  # action_type = "setMinTarget"
  # min_target_capacity = 1
  // ==================================

  // === ADJUSTMENT ===================
  # action_type = "percentageAdjustment"
  # action_type = "adjustment"
  # adjustment = "MAX(5,10)"
  // ==================================

  // === UPDATE CAPACITY ==============
  action_type = "updateCapacity"
  minimum = 0
  maximum = 10
  target = 5
  // ==================================

  }]
 // ----------------------------------------
`

const testCoreScalingDownPolicy_EmptyFields = `
 // --- CORE SCALE DOWN POLICY -------------
 // ----------------------------------------
`

// endregion

// region Create New Cluster Optional Fields

func TestAccSpotinstMRScalerAWSNewCluster_OptionalFields(t *testing.T) {
	scalerName := "mrscaler-new-cluster-optional-fields"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testNewClusterOptionalFields_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "additional_primary_security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_primary_security_groups.0", "sg-f2f94288"),
					resource.TestCheckResourceAttr(resourceName, "additional_replica_security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_replica_security_groups.0", "sg-8cfb40f6"),
					resource.TestCheckResourceAttr(resourceName, "applications.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "applications.1312668776.args.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "applications.1312668776.name", "Ganglia"),
					resource.TestCheckResourceAttr(resourceName, "applications.1312668776.version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "applications.1485771378.args.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "applications.1485771378.name", "Hadoop"),
					resource.TestCheckResourceAttr(resourceName, "applications.1663287870.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "applications.1663287870.args.0", "fake"),
					resource.TestCheckResourceAttr(resourceName, "applications.1663287870.args.1", "args"),
					resource.TestCheckResourceAttr(resourceName, "applications.1663287870.name", "Pig"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_actions_file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_actions_file.3624379019.bucket", "terraform-emr-test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_actions_file.3624379019.key", "bootstrap-actions-file.json"),
					resource.TestCheckResourceAttr(resourceName, "configurations_file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations_file.3778944667.bucket", "terraform-emr-test"),
					resource.TestCheckResourceAttr(resourceName, "configurations_file.3778944667.key", "configurations.json"),
					resource.TestCheckResourceAttr(resourceName, "custom_ami_id", "ami-07b8d9983434da94e"),
					resource.TestCheckResourceAttr(resourceName, "ec2_key_name", "test-key"),
					resource.TestCheckResourceAttr(resourceName, "managed_primary_security_group", "sg-8cfb40f6"),
					resource.TestCheckResourceAttr(resourceName, "managed_replica_security_group", "sg-f2f94288"),
					resource.TestCheckResourceAttr(resourceName, "repo_upgrade_on_boot", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "service_access_security_group", "access-example"),
					resource.TestCheckResourceAttr(resourceName, "steps_file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps_file.1986180246.bucket", "terraform-emr-test"),
					resource.TestCheckResourceAttr(resourceName, "steps_file.1986180246.key", "additional-steps-test.json"),
				),
			},
		},
	})
}

const testNewClusterOptionalFields_Create = `
// --- OPTIONAL FIELDS --------------------
//  ebs_root_volume_size = 10
  custom_ami_id        = "ami-07b8d9983434da94e"
  repo_upgrade_on_boot = "NONE"
  ec2_key_name         = "test-key"

  managed_primary_security_group = "sg-8cfb40f6"
  managed_replica_security_group = "sg-f2f94288"
  service_access_security_group  = "access-example"

  additional_primary_security_groups = ["sg-f2f94288"]
  additional_replica_security_groups = ["sg-8cfb40f6"]

  applications = [
    {
      name = "Ganglia"
      version = "1.0"
    },
    {
      name = "Hadoop"
    },
    {
      name = "Pig"
      args = ["fake", "args"]
    }
  ]

  steps_file = {
    bucket = "terraform-emr-test"
    key = "additional-steps-test.json"
  }

  configurations_file = {
    bucket = "terraform-emr-test"
    key = "configurations.json"
  }

  bootstrap_actions_file = {
    bucket = "terraform-emr-test"
    key = "bootstrap-actions-file.json"
  }
// ----------------------------------------
`

// endregion

// region MRScaler: Scheduled Tasks
func TestAccSpotinstMRScalerAWS_ScheduledTask(t *testing.T) {
	scalerName := "mrscaler-scheduled-task"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t, "aws") },
		Providers:     TestAccProviders,
		CheckDestroy:  testMRScalerAWSDestroy,
		IDRefreshName: resourceName,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testMRScalerAWSScheduledTask_Create,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.1818405131.cron", "* * * * *"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.1818405131.desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.1818405131.instance_group_type", "task"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.1818405131.is_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.1818405131.max_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.1818405131.min_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.1818405131.task_type", "setCapacity"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testMRScalerAWSScheduledTask_Update,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.4264699508.cron", "* * 8 * *"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.4264699508.desired_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.4264699508.instance_group_type", "task"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.4264699508.is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.4264699508.max_capacity", "3"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.4264699508.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.4264699508.task_type", "setCapacity"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testMRScalerAWSScheduledTask_EmptyFields,
					newCluster:     true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "scheduled_task.#", "0"),
				),
			},
		},
	})
}

const testMRScalerAWSScheduledTask_Create = `
 // --- SCHEDULED TASK ------------------
  scheduled_task = [{
    is_enabled = false
    cron = "* * * * *"
    task_type = "setCapacity"
    instance_group_type = "task"
    min_capacity = 0
    max_capacity = 2
    desired_capacity = 1
  }]
 // -------------------------------------
`

const testMRScalerAWSScheduledTask_Update = `
 // --- SCHEDULED TASK ------------------
  scheduled_task = [{
    is_enabled = true
    cron = "* * 8 * *"
    task_type = "setCapacity"
    instance_group_type = "task"
    min_capacity = 1
    max_capacity = 3
    desired_capacity = 2
  }]
 // -------------------------------------
`

const testMRScalerAWSScheduledTask_EmptyFields = `
 // --- SCHEDULED TASK ------------------
 // -------------------------------------
`

// endregion

/*************************************
 *            Cloned Cluster		 *
 *************************************/

// region MRScalerAWSCloned: Baseline
func TestAccSpotinstMRScalerAWSCloned_Baseline(t *testing.T) {
	scalerName := "mrscaler-cloned-baseline"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:           scalerName,
					clonedCluster:        true,
					updateBaselineFields: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
				),
			},
		},
	})
}

const testMRScalerAWSBaselineCloned_Create = `
resource "` + string(commons.MRScalerAWSResourceName) + `" "%v" {
 provider = "%v"

 name               = "%v"
 description        = "test cloning a cluster"
 availability_zones = ["us-west-2b:subnet-1ba25052"]
 strategy           = "%v"
 region             = "us-west-2"
 cluster_id         = "%v"

 %v
 %v
 %v
 %v
 %v
}
`

const testMRScalerAWSBaselineCloned_Update = `
resource "` + string(commons.MRScalerAWSResourceName) + `" "%v" {
 provider = "%v"

 name               = "%v"
 description        = "test updating a cloned cluster"
 availability_zones = ["us-west-2b:subnet-1ba25052"]
 strategy           = "%v"
 region             = "us-west-2"
 cluster_id         = "%v"

 %v
 %v
 %v
 %v
 %v
}
`

// endregion

// region Cloned: Strategy

func TestAccSpotinstMRScalerAWSCloned_Strategy(t *testing.T) {
	scalerName := "mrscaler-cloned-strategy"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					strategyConfig: testMRScalerAWSStrategy_Create,
					clonedCluster:  true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-5.17.0"),
					//resource.TestCheckResourceAttr(resourceName, "retries", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout", "15"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout_action", "terminate"),
				),
			},
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					strategyConfig: testMRScalerAWSStrategy_Update,
					clonedCluster:  true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-5.17.0"),
					//resource.TestCheckResourceAttr(resourceName, "retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout", "20"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_timeout.0.timeout_action", "terminate"),
				),
			},
		},
	})
}

// endregion

// region Cloned Instance Groups: Master Group

func TestAccSpotinstMRScalerAWSCloned_MasterGroup(t *testing.T) {
	scalerName := "mrscaler-cloned-master-group"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					masterGroup:   testMRScalerAWSMasterGroup_Create,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "master_instance_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "master_lifecycle", "SPOT"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_optimized", "true"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_block_device.1008334328.volumes_per_instance", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_block_device.1008334328.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "master_ebs_block_device.1008334328.size_in_gb", "30"),
				),
			},
		},
	})
}

// endregion

// region Cloned Instance Groups: Task Group

func TestAccSpotinstMRScalerAWSCloned_TaskGroup(t *testing.T) {
	scalerName := "mrscaler-cloned-task-group"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					taskGroup:     testMRScalerAWSTaskGroup_Create,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_min_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "task_max_size", "30"),
					resource.TestCheckResourceAttr(resourceName, "task_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_lifecycle", "SPOT"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.3329897523.volumes_per_instance", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.3329897523.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.3329897523.size_in_gb", "40"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					taskGroup:     testMRScalerAWSTaskGroup_Update,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_desired_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_lifecycle", "SPOT"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_optimized", "true"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.1008334328.volumes_per_instance", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.1008334328.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "task_ebs_block_device.1008334328.size_in_gb", "30"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					taskGroup:     testMRScalerAWSTaskGroup_EmptyFields,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "task_min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_max_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_desired_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "task_lifecycle", "SPOT"),
				),
			},
		},
	})
}

// endregion

// region Cloned: Instance Groups: Core Group

func TestAccSpotinstMRScalerAWSCloned_CoreGroup(t *testing.T) {
	scalerName := "mrscaler-cloned-core-group"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					coreGroup:     testMRScalerAWSCoreGroup_Create,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_lifecycle", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.3329897523.volumes_per_instance", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.3329897523.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.3329897523.size_in_gb", "40"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					coreGroup:     testMRScalerAWSCoreGroup_Update,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_lifecycle", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_optimized", "true"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.1008334328.volumes_per_instance", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.1008334328.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "core_ebs_block_device.1008334328.size_in_gb", "30"),
				),
			},
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:    scalerName,
					coreGroup:     testMRScalerAWSCoreGroup_EmptyFields,
					clonedCluster: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "core_min_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_max_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_desired_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.0", "c3.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_types.1", "c4.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "core_lifecycle", "ON_DEMAND"),
				),
			},
		},
	})
}

// endregion

// region Cloned: Tags

func TestAccSpotinstMRScalerAWSCloned_Tags(t *testing.T) {
	scalerName := "mrscaler-cloned-tags"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testMRScalerAWSTags_Create,
					clonedCluster:  true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.664003903.key", "Creator"),
					resource.TestCheckResourceAttr(resourceName, "tags.664003903.value", "Terraform"),
				),
			},
		},
	})
}

// endregion

// region Create Cloned Cluster Optional Fields

func TestAccSpotinstMRScalerAWSCloned_OptionalFields(t *testing.T) {
	scalerName := "mrscaler-cloned-cluster-optional-fields"
	resourceName := createMRScalerAWSResourceName(scalerName)

	var scaler mrscaler.Scaler
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t, "aws") },
		Providers:    TestAccProviders,
		CheckDestroy: testMRScalerAWSDestroy,

		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config: createMRScalerAWSTerraform(&MRScalerAWSConfigMetaData{
					scalerName:     scalerName,
					fieldsToAppend: testClonedOptionalFields_Create,
					clonedCluster:  true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckMRScalerAWSExists(&scaler, resourceName),
					testCheckMRScalerAWSAttributes(&scaler, scalerName),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_actions_file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_actions_file.3624379019.bucket", "terraform-emr-test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_actions_file.3624379019.key", "bootstrap-actions-file.json"),
					resource.TestCheckResourceAttr(resourceName, "configurations_file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations_file.3778944667.bucket", "terraform-emr-test"),
					resource.TestCheckResourceAttr(resourceName, "configurations_file.3778944667.key", "configurations.json"),
					resource.TestCheckResourceAttr(resourceName, "steps_file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps_file.1986180246.bucket", "terraform-emr-test"),
					resource.TestCheckResourceAttr(resourceName, "steps_file.1986180246.key", "additional-steps-test.json"),
				),
			},
		},
	})
}

const testClonedOptionalFields_Create = `
// --- OPTIONAL FIELDS --------------------
//  ebs_root_volume_size = 10

  steps_file = {
    bucket = "terraform-emr-test"
    key = "additional-steps-test.json"
  }

  configurations_file = {
    bucket = "terraform-emr-test"
    key = "configurations.json"
  }

  bootstrap_actions_file = {
    bucket = "terraform-emr-test"
    key = "bootstrap-actions-file.json"
  }
// ----------------------------------------
`

// endregion
