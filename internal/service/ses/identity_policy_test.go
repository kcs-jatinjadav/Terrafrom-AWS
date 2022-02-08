package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
)

func TestAccSESIdentityPolicy_basic(t *testing.T) {
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyIdentityDomainConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(resourceName),
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

func TestAccSESIdentityPolicy_Identity_email(t *testing.T) {
	emailPrefix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := fmt.Sprintf("%s@%s", emailPrefix, acctest.RandomDomainName())
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyIdentityEmailConfig(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(resourceName),
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

func TestAccSESIdentityPolicy_policy(t *testing.T) {
	domain := acctest.RandomDomainName()
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyPolicy1Config(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(resourceName),
				),
			},
			{
				Config: testAccIdentityPolicyPolicy2Config(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(resourceName),
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

func TestAccSESIdentityPolicy_ignoreEquivalent(t *testing.T) {
	domain := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_identity_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityPolicyEquivalentConfig(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityPolicyExists(resourceName),
				),
			},
			{
				Config:   testAccIdentityPolicyEquivalent2Config(rName, domain),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckIdentityPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_identity_policy" {
			continue
		}

		identityARN, policyName, err := tfses.IdentityPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &ses.GetIdentityPoliciesInput{
			Identity:    aws.String(identityARN),
			PolicyNames: aws.StringSlice([]string{policyName}),
		}

		output, err := conn.GetIdentityPolicies(input)

		if err != nil {
			return err
		}

		if output != nil && len(output.Policies) > 0 && aws.StringValue(output.Policies[policyName]) != "" {
			return fmt.Errorf("SES Identity (%s) Policy (%s) still exists", identityARN, policyName)
		}
	}

	return nil
}

func testAccCheckIdentityPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("SES Identity Policy not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Identity Policy ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		identityARN, policyName, err := tfses.IdentityPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &ses.GetIdentityPoliciesInput{
			Identity:    aws.String(identityARN),
			PolicyNames: aws.StringSlice([]string{policyName}),
		}

		output, err := conn.GetIdentityPolicies(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.Policies) == 0 {
			return fmt.Errorf("SES Identity (%s) Policy (%s) not found", identityARN, policyName)
		}

		return nil
	}
}

func testAccIdentityPolicyIdentityDomainConfig(domain string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccIdentityPolicyIdentityEmailConfig(email string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_email_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_email_identity" "test" {
  email = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_email_identity.test.email
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, email)
}

func testAccIdentityPolicyPolicy1Config(domain string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccIdentityPolicyPolicy2Config(domain string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.test.arn]

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = "test"
  policy   = data.aws_iam_policy_document.test.json
}
`, domain)
}

func testAccIdentityPolicyEquivalentConfig(rName, domain string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = %[2]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[2]q
    Statement = [{
      Sid    = %[2]q
      Effect = "Allow"
      Principal = {
        AWS = [data.aws_caller_identity.current.account_id]
      }
      Action = [
        "SES:SendEmail",
        "SES:SendRawEmail",
      ]
      Resource = [aws_ses_domain_identity.test.arn]
    }]
  })
}
`, domain, rName)
}

func testAccIdentityPolicyEquivalent2Config(rName, domain string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_identity_policy" "test" {
  identity = aws_ses_domain_identity.test.arn
  name     = %[2]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[2]q
    Statement = [{
      Sid    = %[2]q
      Effect = "Allow"
      Principal = {
        AWS = data.aws_caller_identity.current.account_id
      }
      Action = [
        "SES:SendRawEmail",
        "SES:SendEmail",
      ]
      Resource = aws_ses_domain_identity.test.arn
    }]
  })
}
`, domain, rName)
}