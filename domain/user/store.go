// Copyright 2016 Documize Inc. <legal@documize.com>. All rights reserved.
//
// This software (Documize Community Edition) is licensed under
// GNU AGPL v3 http://www.gnu.org/licenses/agpl-3.0.en.html
//
// You can operate outside the AGPL restrictions by purchasing
// Documize Enterprise Edition and obtaining a commercial license
// by contacting <sales@documize.com>.
//
// https://documize.com

package user

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/documize/community/core/streamutil"
	"github.com/documize/community/domain"
	"github.com/pkg/errors"
)

// Add adds the given user record to the user table.
func Add(s domain.StoreContext, u User) (err error) {
	u.Created = time.Now().UTC()
	u.Revised = time.Now().UTC()

	stmt, err := s.Context.Transaction.Preparex("INSERT INTO user (refid, firstname, lastname, email, initials, password, salt, reset, created, revised) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, "prepare user insert")
		return
	}

	_, err = stmt.Exec(u.RefID, u.Firstname, u.Lastname, strings.ToLower(u.Email), u.Initials, u.Password, u.Salt, "", u.Created, u.Revised)
	if err != nil {
		err = errors.Wrap(err, "execute user insert")
		return
	}

	return
}

// Get returns the user record for the given id.
func Get(s domain.StoreContext, id string) (u User, err error) {
	stmt, err := s.Runtime.Db.Preparex("SELECT id, refid, firstname, lastname, email, initials, global, password, salt, reset, created, revised FROM user WHERE refid=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("unable to prepare select for user %s", id))
		return
	}

	err = stmt.Get(&u, id)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("unable to execute select for user %s", id))
		return
	}

	return
}

// GetByDomain matches user by email and domain.
func GetByDomain(s domain.StoreContext, domain, email string) (u User, err error) {
	email = strings.TrimSpace(strings.ToLower(email))

	stmt, err := s.Runtime.Db.Preparex("SELECT u.id, u.refid, u.firstname, u.lastname, u.email, u.initials, u.global, u.password, u.salt, u.reset, u.created, u.revised FROM user u, account a, organization o WHERE TRIM(LOWER(u.email))=? AND u.refid=a.userid AND a.orgid=o.refid AND TRIM(LOWER(o.domain))=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Unable to prepare GetUserByDomain %s %s", domain, email))
		return
	}

	err = stmt.Get(&u, email, domain)
	if err != nil && err != sql.ErrNoRows {
		err = errors.Wrap(err, fmt.Sprintf("Unable to execute GetUserByDomain %s %s", domain, email))
		return
	}

	return
}

// GetByEmail returns a single row match on email.
func GetByEmail(s domain.StoreContext, email string) (u User, err error) {
	email = strings.TrimSpace(strings.ToLower(email))

	stmt, err := s.Runtime.Db.Preparex("SELECT id, refid, firstname, lastname, email, initials, global, password, salt, reset, created, revised FROM user WHERE TRIM(LOWER(email))=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("prepare select user by email %s", email))
		return
	}

	err = stmt.Get(&u, email)
	if err != nil && err != sql.ErrNoRows {
		err = errors.Wrap(err, fmt.Sprintf("execute select user by email %s", email))
		return
	}

	return
}

// GetByToken returns a user record given a reset token value.
func GetByToken(s domain.StoreContext, token string) (u User, err error) {
	stmt, err := s.Runtime.Db.Preparex("SELECT  id, refid, firstname, lastname, email, initials, global, password, salt, reset, created, revised FROM user WHERE reset=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("prepare user select by token %s", token))
		return
	}

	err = stmt.Get(&u, token)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("execute user select by token %s", token))
		return
	}

	return
}

// GetBySerial is used to retrieve a user via their temporary password salt value!
// This occurs when we you share a folder with a new user and they have to complete
// the onboarding process.
func GetBySerial(s domain.StoreContext, serial string) (u User, err error) {
	stmt, err := s.Runtime.Db.Preparex("SELECT id, refid, firstname, lastname, email, initials, global, password, salt, reset, created, revised FROM user WHERE salt=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("prepare user select by serial %s", serial))
		return
	}

	err = stmt.Get(&u, serial)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("execute user select by serial %s", serial))
		return
	}

	return
}

// GetActiveUsersForOrganization returns a slice containing of active user records for the organization
// identified in the Persister.
func GetActiveUsersForOrganization(s domain.StoreContext) (u []User, err error) {
	err = s.Runtime.Db.Select(&u,
		`SELECT u.id, u.refid, u.firstname, u.lastname, u.email, u.initials, u.password, u.salt, u.reset, u.created, u.revised
		FROM user u
		WHERE u.refid IN (SELECT userid FROM account WHERE orgid = ? AND active=1) ORDER BY u.firstname,u.lastname`,
		s.Context.OrgID)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("get active users by org %s", s.Context.OrgID))
		return
	}

	return
}

// GetUsersForOrganization returns a slice containing all of the user records for the organizaiton
// identified in the Persister.
func GetUsersForOrganization(s domain.StoreContext) (u []User, err error) {
	err = s.Runtime.Db.Select(&u,
		"SELECT id, refid, firstname, lastname, email, initials, password, salt, reset, created, revised FROM user WHERE refid IN (SELECT userid FROM account where orgid = ?) ORDER BY firstname,lastname", s.Context.OrgID)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf(" get users for org %s", s.Context.OrgID))
		return
	}

	return
}

// GetSpaceUsers returns a slice containing all user records for given folder.
func GetSpaceUsers(s domain.StoreContext, folderID string) (u []User, err error) {
	err = s.Runtime.Db.Select(&u,
		`SELECT u.id, u.refid, u.firstname, u.lastname, u.email, u.initials, u.password, u.salt, u.reset, u.created, u.revised 
		FROM user u, account a
		WHERE u.refid IN (SELECT userid from labelrole WHERE orgid=? AND labelid=?) 
		AND a.orgid=? AND u.refid = a.userid AND a.active=1
		ORDER BY u.firstname, u.lastname`,
		s.Context.OrgID, folderID, s.Context.OrgID)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("get space users for org %s", s.Context.OrgID))
		return
	}

	return
}

// UpdateUser updates the user table using the given replacement user record.
func UpdateUser(s domain.StoreContext, u User) (err error) {
	u.Revised = time.Now().UTC()
	u.Email = strings.ToLower(u.Email)

	stmt, err := s.Context.Transaction.PrepareNamed(
		"UPDATE user SET firstname=:firstname, lastname=:lastname, email=:email, revised=:revised, initials=:initials WHERE refid=:refid")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("prepare user update %s", u.RefID))
		return
	}

	_, err = stmt.Exec(&u)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("execute user update %s", u.RefID))
		return
	}

	return
}

// UpdateUserPassword updates a user record with new password and salt values.
func UpdateUserPassword(s domain.StoreContext, userID, salt, password string) (err error) {
	stmt, err := s.Context.Transaction.Preparex("UPDATE user SET salt=?, password=?, reset='' WHERE refid=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, "prepare user update")
		return
	}

	_, err = stmt.Exec(salt, password, userID)
	if err != nil {
		err = errors.Wrap(err, "execute user update")
		return
	}

	return
}

// DeactiveUser deletes the account record for the given userID and persister.Context.OrgID.
func DeactiveUser(s domain.StoreContext, userID string) (err error) {
	stmt, err := s.Context.Transaction.Preparex("DELETE FROM account WHERE userid=? and orgid=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, "prepare user deactivation")
		return
	}

	_, err = stmt.Exec(userID, s.Context.OrgID)

	if err != nil {
		err = errors.Wrap(err, "execute user deactivation")
		return
	}

	return
}

// ForgotUserPassword sets the password to '' and the reset field to token, for a user identified by email.
func ForgotUserPassword(s domain.StoreContext, email, token string) (err error) {
	stmt, err := s.Context.Transaction.Preparex("UPDATE user SET reset=?, password='' WHERE LOWER(email)=?")
	defer streamutil.Close(stmt)

	if err != nil {
		err = errors.Wrap(err, "prepare password reset")
		return
	}

	_, err = stmt.Exec(token, strings.ToLower(email))
	if err != nil {
		err = errors.Wrap(err, "execute password reset")
		return
	}

	return
}

// CountActiveUsers returns the number of active users in the system.
func CountActiveUsers(s domain.StoreContext) (c int) {
	row := s.Runtime.Db.QueryRow("SELECT count(*) FROM user u WHERE u.refid IN (SELECT userid FROM account WHERE active=1)")

	err := row.Scan(&c)

	if err == sql.ErrNoRows {
		return 0
	}

	if err != nil && err != sql.ErrNoRows {
		s.Runtime.Log.Error("CountActiveUsers", err)
		return 0
	}

	return
}