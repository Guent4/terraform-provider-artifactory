package artifactory

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLdapSetting_full(t *testing.T) {
	const LdapSettingTemplateFull = `
resource "artifactory_ldap_setting" "ldaptest" {
	key = "ldaptest"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=Peo *ple, uid={0}"
	email_attribute = "mail_attr"
	search_sub_tree = true
	search_filter = "(uid={0})"
	search_base = "ou=users|ou=people"
	manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
	manager_password = "testmgrpaswd"
}`

	const LdapSettingTemplateUpdate = `
resource "artifactory_ldap_setting" "ldaptest" {
	key = "ldaptest"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "mail_attr"
	search_sub_tree = true
	search_filter = "(uid={0})"
	search_base = "ou=users"
	manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
	manager_password = "testmgrpaswd"
}`
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccLdapSettingDestroy("ldaptest"),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: LdapSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "ldap_url", "ldap://ldaptestldap"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "user_dn_pattern", "ou=Peo *ple, uid={0}"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "email_attribute", "mail_attr"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "search_sub_tree", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "search_filter", "(uid={0})"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "search_base", "ou=users|ou=people"),
				),
			},
			{
				Config: LdapSettingTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "email_attribute", "mail_attr"),
				),
			},
		},
	})
}

func TestAccLdapSetting_emailAttribute(t *testing.T) {
	const LdapSettingTemplateNoEmailAttr = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=People, uid={0}"
}`

	const LdapSettingTemplateEmailAttrBlank = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=People, uid={0}"
	email_attribute = ""
}`

	const LdapSettingTemplateEmailAttrUpd1 = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "mail"
}`

	const LdapSettingTemplateEmailAttrUpd2 = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "mail_attr"
}`

	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccLdapSettingDestroy("ldaptestemailattr"),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: LdapSettingTemplateNoEmailAttr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "email_attribute", "mail"),
				),
			},
			{
				Config: LdapSettingTemplateEmailAttrBlank,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "email_attribute", "mail"),
				),
			},
			{
				Config: LdapSettingTemplateEmailAttrUpd1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "email_attribute", "mail"),
				),
			},
			{
				Config: LdapSettingTemplateEmailAttrUpd2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptestemailattr", "email_attribute", "mail_attr"),
				),
			},
		},
	})
}

func testAccLdapSettingDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := testAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources["artifactory_ldap_setting."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		ldapConfigs := &XmlLdapConfig{}

		response, err := client.R().SetResult(&ldapConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		for _, iterLdapSetting := range ldapConfigs.Security.LdapSettings.LdapSettingArr {
			if iterLdapSetting.Key == id {
				return fmt.Errorf("error: LdapSetting with key: " + id + " still exists.")
			}
		}
		return nil
	}
}
