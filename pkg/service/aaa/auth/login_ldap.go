// Copyright 2022 The kubegems.io Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"kubegems.io/kubegems/pkg/i18n"
	"kubegems.io/kubegems/pkg/log"
)

type LdapLoginUtils struct {
	Name         string `yaml:"name" json:"name"`
	Vendor       string `json:"vendor"`
	LdapAddr     string `yaml:"addr" json:"ldapaddr"`
	BaseDN       string `yaml:"basedn" json:"basedn"`
	EnableTLS    bool   `json:"enableTLS"`
	Filter       string `json:"filter"`
	BindUsername string `yaml:"binduser" json:"binduser"`
	BindPassword string `yaml:"bindpass" json:"password"`
}

func (ut *LdapLoginUtils) GetName() string {
	return ut.Name
}

func (ut *LdapLoginUtils) LoginAddr() string {
	return DefaultLoginURL
}

func (ut *LdapLoginUtils) GetUserInfo(ctx context.Context, cred *Credential) (ret *UserInfo, err error) {
	if !ut.ValidateCredential(cred) {
		return nil, i18n.Errorf(ctx, "invalid credential")
	}
	var ldapConn *ldap.Conn
	ldap.DefaultTimeout = time.Second * 5
	if strings.HasPrefix(ut.LdapAddr, "ldap") {
		ldapConn, err = ldap.DialURL(
			ut.LdapAddr,
			ldap.DialWithDialer(&net.Dialer{Timeout: time.Second * 5}),
		)
	} else {
		ldapConn, err = ldap.Dial("tcp", ut.LdapAddr)
	}
	defer ldapConn.Close()
	if err != nil {
		log.Error(err, "connect to ldap server failed")
		return nil, i18n.Error(ctx, "failed to connect ldap server")
	}

	if ut.EnableTLS {
		if err = ldapConn.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
			log.Error(err, "failed to connect ldap server with tls")
			return nil, i18n.Error(ctx, "failed to connect ldap server with tls")
		}
	}

	if err = ldapConn.Bind(ut.BindUsername, ut.BindPassword); err != nil {
		log.Error(err, "failed to connect server with tls")
		return nil, i18n.Error(ctx, "failed to connect ldap server with tls")
	}

	searchRequest := ldap.NewSearchRequest(
		ut.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		fmt.Sprintf("(cn=%s)", cred.Username),
		[]string{"mail"},
		nil,
	)
	result, err := ldapConn.Search(searchRequest)
	if err != nil {
		log.Error(err, "search user in ldap failed")
		return nil, i18n.Error(ctx, "failed to get userinfo from ldap")
	}
	if len(result.Entries) != 1 {
		log.Error(fmt.Errorf("more than one search result returnd"), "username", cred.Username)
		return nil, i18n.Error(ctx, "failed to get userinfo from ldap, more than one result")
	}
	uinfo := UserInfo{}
	info := result.Entries[0]
	uinfo.Username = cred.Username
	uinfo.Vendor = ut.Vendor
	mailstr := info.GetAttributeValue("mail")
	emailstr := info.GetAttributeValue("email")
	if emailstr != "" {
		uinfo.Email = emailstr
	} else {
		uinfo.Email = mailstr
	}
	uinfo.Source = cred.Source
	ret = &uinfo
	return
}

func (ut *LdapLoginUtils) ValidateCredential(cred *Credential) bool {
	userdn := fmt.Sprintf("cn=%s,%s", cred.Username, ut.BaseDN)
	req := ldap.NewSimpleBindRequest(userdn, cred.Password, nil)

	var (
		ldapConn *ldap.Conn
		err      error
	)
	ldap.DefaultTimeout = time.Second * 5
	if strings.HasPrefix(ut.LdapAddr, "ldap") {
		ldapConn, err = ldap.DialURL(
			ut.LdapAddr,
			ldap.DialWithDialer(&net.Dialer{Timeout: time.Second * 5}),
		)
	} else {
		ldapConn, err = ldap.Dial("tcp", ut.LdapAddr)
	}
	defer ldapConn.Close()
	if err != nil {
		log.Error(err, "connect to ldap server failed")
		return false
	}

	if ut.EnableTLS {
		if err := ldapConn.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
			log.Error(err, "connect to ldap server with tls failed")
			return false
		}
	}
	_, err = ldapConn.SimpleBind(req)
	if err != nil {
		log.Error(err, "faield to login with ldap", "enableTLS", ut.EnableTLS, "username", cred.Username, "source", cred.Source)
		return false
	}
	return true
}
