package imagebuilder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceDistributionConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDistributionConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distribution": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ami_distribution_configuration": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ami_tags": tftags.TagsSchemaComputed(),
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"launch_permission": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"user_groups": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"user_ids": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"target_account_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"container_distribution_configuration": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_tags": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"target_repository": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"service": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"launch_template_configuration": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"launch_template_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"license_configuration_arns": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceDistributionConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetDistributionConfigurationInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.DistributionConfigurationArn = aws.String(v.(string))
	}

	output, err := conn.GetDistributionConfiguration(input)

	if err != nil {
		return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.DistributionConfiguration == nil {
		return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): empty response", d.Id())
	}

	distributionConfiguration := output.DistributionConfiguration

	d.SetId(aws.StringValue(distributionConfiguration.Arn))
	d.Set("arn", distributionConfiguration.Arn)
	d.Set("date_created", distributionConfiguration.DateCreated)
	d.Set("date_updated", distributionConfiguration.DateUpdated)
	d.Set("description", distributionConfiguration.Description)
	d.Set("distribution", flattenDistributions(distributionConfiguration.Distributions))
	d.Set("name", distributionConfiguration.Name)
	d.Set("tags", KeyValueTags(distributionConfiguration.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	return nil
}
