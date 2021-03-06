package s3_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestBucketACLParseResourceID(t *testing.T) {
	testCases := []struct {
		TestName            string
		InputID             string
		ExpectError         bool
		ExpectedACL         string
		ExpectedBucket      string
		ExpectedBucketOwner string
	}{
		{
			TestName:    "empty ID",
			InputID:     "",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket and account ID format with slash separator",
			InputID:     "test/123456789012",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket, account ID, and ACL format with slash separators",
			InputID:     "test/123456789012/private",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket, account ID, and ACL format with mixed separators",
			InputID:     "test/123456789012,private",
			ExpectError: true,
		},
		{
			TestName:    "incorrect bucket, ACL, and account ID format",
			InputID:     "test,private,123456789012",
			ExpectError: true,
		},
		{
			TestName:            "valid ID with bucket",
			InputID:             tfs3.BucketACLCreateResourceID("example", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example-bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example-bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dot and hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dots, hyphen, and numbers",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and acl",
			InputID:             tfs3.BucketACLCreateResourceID("example", "", s3.BucketCannedACLPrivate),
			ExpectedACL:         s3.BucketCannedACLPrivate,
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("example", "", s3.BucketCannedACLPublicReadWrite),
			ExpectedACL:         s3.BucketCannedACLPublicReadWrite,
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and number and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "", s3.BucketCannedACLPublicReadWrite),
			ExpectedACL:         s3.BucketCannedACLPublicReadWrite,
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and bucket owner",
			InputID:             tfs3.BucketACLCreateResourceID("example", "123456789012", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and number and bucket owner",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "123456789012", ""),
			ExpectedACL:         "",
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket, bucket owner, and acl",
			InputID:             tfs3.BucketACLCreateResourceID("example", "123456789012", s3.BucketCannedACLPrivate),
			ExpectedACL:         s3.BucketCannedACLPrivate,
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket, bucket owner, and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("example", "123456789012", s3.BucketCannedACLPublicReadWrite),
			ExpectedACL:         s3.BucketCannedACLPublicReadWrite,
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "123456789012",
		},
		{
			TestName:            "valid ID with bucket that has dot, hyphen, and numbers, bucket owner, and acl that has hyphens",
			InputID:             tfs3.BucketACLCreateResourceID("my-example.bucket.4000", "123456789012", s3.BucketCannedACLPublicReadWrite),
			ExpectedACL:         s3.BucketCannedACLPublicReadWrite,
			ExpectedBucket:      "my-example.bucket.4000",
			ExpectedBucketOwner: "123456789012",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotBucket, gotExpectedBucketOwner, gotAcl, err := tfs3.BucketACLParseResourceID(testCase.InputID)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("unexpected error: %s", err)
			}

			if gotAcl != testCase.ExpectedACL {
				t.Errorf("got ACL %s, expected %s", gotAcl, testCase.ExpectedACL)
			}

			if gotBucket != testCase.ExpectedBucket {
				t.Errorf("got bucket %s, expected %s", gotBucket, testCase.ExpectedBucket)
			}

			if gotExpectedBucketOwner != testCase.ExpectedBucketOwner {
				t.Errorf("got ExpectedBucketOwner %s, expected %s", gotExpectedBucketOwner, testCase.ExpectedBucketOwner)
			}
		})
	}
}

func TestAccS3BucketAcl_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAclBasicConfig(bucketName, s3.BucketCannedACLPrivate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", s3.BucketCannedACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeCanonicalUser,
						"permission":     s3.PermissionFullControl,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketAcl_disappears(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAclBasicConfig(bucketName, s3.BucketCannedACLPrivate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					// Bucket ACL cannot be destroyed, but we can verify Bucket deletion
					// will result in a missing Bucket ACL resource
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucket(), "aws_s3_bucket.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketAcl_updateACL(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAclBasicConfig(bucketName, s3.BucketCannedACLPublicRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", s3.BucketCannedACLPublicRead),
				),
			},
			{
				Config: testAccBucketAclBasicConfig(bucketName, s3.BucketCannedACLPrivate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", s3.BucketCannedACLPrivate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketAcl_updateGrant(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAcl_GrantsConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeCanonicalUser,
						"permission":     s3.PermissionFullControl,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeCanonicalUser,
						"permission":     s3.PermissionWrite,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", "id"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketAcl_GrantsUpdateConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeCanonicalUser,
						"permission":     s3.PermissionRead,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "access_control_policy.0.grant.*.grantee.0.id", "data.aws_canonical_user_id.current", "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeGroup,
						"permission":     s3.PermissionReadAcp,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "access_control_policy.0.grant.*", map[string]*regexp.Regexp{
						"grantee.0.uri": regexp.MustCompile(`http://acs.*/groups/s3/LogDelivery`),
					}),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.owner.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_control_policy.0.owner.0.id", "data.aws_canonical_user_id.current", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketAcl_ACLToGrant(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAclBasicConfig(bucketName, s3.BucketCannedACLPrivate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", s3.BucketCannedACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
				),
			},
			{
				Config: testAccBucketAcl_GrantsConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketAcl_grantToACL(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAcl_GrantsConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.0.grant.#", "2"),
				),
			},
			{
				Config: testAccBucketAclBasicConfig(bucketName, s3.BucketCannedACLPrivate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAclExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", s3.BucketCannedACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "access_control_policy.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBucketAclExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		bucket, expectedBucketOwner, _, err := tfs3.BucketACLParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketAclInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketAcl(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.Grants) == 0 || output.Owner == nil {
			return fmt.Errorf("S3 bucket ACL %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketAclBasicConfig(rName, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = %[2]q
}
`, rName, acl)
}

func testAccBucketAcl_GrantsConfig(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "WRITE"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
}
`, bucketName)
}

func testAccBucketAcl_GrantsUpdateConfig(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "READ"
    }

    grant {
      grantee {
        type = "Group"
        uri  = "http://acs.${data.aws_partition.current.dns_suffix}/groups/s3/LogDelivery"
      }
      permission = "READ_ACP"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
}
`, bucketName)
}
