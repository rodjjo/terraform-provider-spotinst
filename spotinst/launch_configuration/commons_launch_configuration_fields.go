package launch_configuration

import (
	"fmt"
	"encoding/base64"
	"crypto/sha1"
	"encoding/hex"

	"github.com/terraform-providers/terraform-provider-spotinst/spotinst/commons"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"regexp"
)

//-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-
//            Setup
//-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-
func SetupSpotinstLaunchConfigurationResource() {
	fields := make(map[commons.FieldName]*commons.GenericField)
	var readFailurePattern = "launch configuration failed reading field %s - %#v"

	//if lspec.ShutdownScript != nil && spotinst.StringValue(lspec.ShutdownScript) != "" {
	//	decodedShutdownScript, _ := base64.StdEncoding.DecodeString(spotinst.StringValue(lspec.ShutdownScript))
	//	result["shutdown_script"] = string(decodedShutdownScript)
	//} else {
	//	result["shutdown_script"] = ""
	//}

	fields[ImageId] = commons.NewGenericField(
		ImageId,
		&schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification
			value := spotinst.StringValue(lc.ImageID)
			if err := resourceData.Set(string(ImageId), value); err != nil {
				return fmt.Errorf(readFailurePattern, string(ImageId), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(ImageId)).(string); ok && v != "" {
				elastigroup.Compute.LaunchSpecification.SetImageId(spotinst.String(v))
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(ImageId)).(string); ok && v != "" {
				elastigroup.Compute.LaunchSpecification.SetImageId(spotinst.String(v))
			}
			return nil
		},
		nil,
	)

	fields[IamInstanceProfile] = commons.NewGenericField(
		IamInstanceProfile,
		&schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification

			var value = ""
			if lc.IAMInstanceProfile != nil {
				if lc.IAMInstanceProfile.Arn != nil {
					value = spotinst.StringValue(lc.IAMInstanceProfile.Arn)
				} else {
					value = spotinst.StringValue(lc.IAMInstanceProfile.Name)
				}
			}

			if err := resourceData.Set(string(IamInstanceProfile), value); err != nil {
				return fmt.Errorf(readFailurePattern, string(IamInstanceProfile), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(IamInstanceProfile)).(string); ok && v != "" {
				iamInstanceProf := &aws.IAMInstanceProfile{}
				if InstanceProfileArnRegex.MatchString(v) {
					iamInstanceProf.SetArn(spotinst.String(v))
				} else {
					iamInstanceProf.SetName(spotinst.String(v))
				}
				elastigroup.Compute.LaunchSpecification.SetIAMInstanceProfile(iamInstanceProf)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(IamInstanceProfile)).(string); ok && v != "" {
				iamInstanceProf := &aws.IAMInstanceProfile{}
				if InstanceProfileArnRegex.MatchString(v) {
					iamInstanceProf.SetArn(spotinst.String(v))
				} else {
					iamInstanceProf.SetName(spotinst.String(v))
				}
				elastigroup.Compute.LaunchSpecification.SetIAMInstanceProfile(iamInstanceProf)
			} else {
				elastigroup.Compute.LaunchSpecification.SetIAMInstanceProfile(nil)
			}
			return nil
		},
		nil,
	)

	fields[KeyName] = commons.NewGenericField(
		KeyName,
		&schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification
			value := spotinst.StringValue(lc.KeyPair)
			if err := resourceData.Set(string(KeyName), value); err != nil {
				return fmt.Errorf(readFailurePattern, string(KeyName), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(KeyName)).(string); ok && v != "" {
				elastigroup.Compute.LaunchSpecification.SetKeyPair(spotinst.String(v))
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(KeyName)).(string); ok && v != "" {
				elastigroup.Compute.LaunchSpecification.SetKeyPair(spotinst.String(v))
			}
			return nil
		},
		nil,
	)

	fields[SecurityGroups] = commons.NewGenericField(
		SecurityGroups,
		&schema.Schema{
			Type:     schema.TypeList,
			Required: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification
			if err := resourceData.Set(string(SecurityGroups), lc.SecurityGroupIDs); err != nil {
				return fmt.Errorf(readFailurePattern, string(SecurityGroups), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(SecurityGroups)).([]interface{}); ok {
				ids := make([]string, len(v))
				for i, j := range v {
					ids[i] = j.(string)
				}
				elastigroup.Compute.LaunchSpecification.SetSecurityGroupIDs(ids)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(SecurityGroups)).([]interface{}); ok {
				ids := make([]string, len(v))
				for i, j := range v {
					ids[i] = j.(string)
				}
				elastigroup.Compute.LaunchSpecification.SetSecurityGroupIDs(ids)
			}
			return nil
		},
		nil,
	)

	fields[UserData] = commons.NewGenericField(
		UserData,
		&schema.Schema{
			Type:      schema.TypeString,
			Optional:  true,
			StateFunc: hexStateFunc,
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification

			var value = ""
			if lc.UserData != nil && spotinst.StringValue(lc.UserData) != "" {
				decodedUserData, _ := base64.StdEncoding.DecodeString(spotinst.StringValue(lc.UserData))
				value = string(decodedUserData)
			}

			if err := resourceData.Set(string(UserData), value); err != nil {
				return fmt.Errorf(readFailurePattern, string(UserData), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(UserData)).(string); ok && v != "" {
				userData := spotinst.String(base64Encode(v))
				elastigroup.Compute.LaunchSpecification.SetUserData(userData)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			var userData *string = nil
			if v, ok := resourceData.Get(string(UserData)).(string); ok && v != "" {
				userData = spotinst.String(base64Encode(v))
			}
			elastigroup.Compute.LaunchSpecification.SetUserData(userData)
			return nil
		},
		nil,
	)

	fields[EnableMonitoring] = commons.NewGenericField(
		EnableMonitoring,
		&schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification
			value := spotinst.BoolValue(lc.Monitoring)
			if err := resourceData.Set(string(EnableMonitoring), value); err != nil {
				return fmt.Errorf(readFailurePattern, string(EnableMonitoring), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(EnableMonitoring)).(bool); ok {
				elastigroup.Compute.LaunchSpecification.SetMonitoring(spotinst.Bool(v))
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(EnableMonitoring)).(bool); ok {
				elastigroup.Compute.LaunchSpecification.SetMonitoring(spotinst.Bool(v))
			}
			return nil
		},
		nil,
	)

	fields[EbsOptimized] = commons.NewGenericField(
		EbsOptimized,
		&schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification
			value := spotinst.BoolValue(lc.EBSOptimized)
			if err := resourceData.Set(string(EbsOptimized), value); err != nil {
				return fmt.Errorf(readFailurePattern, string(EbsOptimized), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(EbsOptimized)).(bool); ok {
				elastigroup.Compute.LaunchSpecification.SetEBSOptimized(spotinst.Bool(v))
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(EbsOptimized)).(bool); ok {
				elastigroup.Compute.LaunchSpecification.SetEBSOptimized(spotinst.Bool(v))
			}
			return nil
		},
		nil,
	)

	fields[PlacementTenancy] = commons.NewGenericField(
		PlacementTenancy,
		&schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			lc := elastigroup.Compute.LaunchSpecification
			value := spotinst.StringValue(lc.Tenancy)
			if err := resourceData.Set(string(PlacementTenancy), value); err != nil {
				return fmt.Errorf(readFailurePattern, string(PlacementTenancy), err)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			if v, ok := resourceData.Get(string(PlacementTenancy)).(string); ok && v != "" {
				tenancy := spotinst.String(v)
				elastigroup.Compute.LaunchSpecification.SetTenancy(tenancy)
			}
			return nil
		},
		func(elastigroup *aws.Group, resourceData *schema.ResourceData, meta interface{}) error {
			var tenancy *string = nil
			if v, ok := resourceData.Get(string(PlacementTenancy)).(string); ok && v != "" {
				tenancy = spotinst.String(v)
			}
			elastigroup.Compute.LaunchSpecification.SetTenancy(tenancy)
			return nil
		},
		nil,
	)

	commons.LaunchConfigurationRepo = commons.NewGenericCachedResource(fields)
}


//-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-
//            Utils
//-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-
var InstanceProfileArnRegex = regexp.MustCompile(`arn:aws:iam::\d{12}:instance-profile/?[a-zA-Z_0-9+=,.@\-_/]+`)

func hexStateFunc(v interface{}) string {
	switch s := v.(type) {
	case string:
		hash := sha1.Sum([]byte(s))
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

// base64Encode encodes data if the input isn't already encoded using
// base64.StdEncoding.EncodeToString. If the input is already base64 encoded,
// return the original input unchanged.
func base64Encode(data string) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return data
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func isBase64Encoded(data string) bool {
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}