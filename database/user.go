// mautrix-groupme - A Matrix-GroupMe puppeting bridge.
// Copyright (C) 2022 Sumner Evans, Karmanyaah Malhotra
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package database

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	log "maunium.net/go/maulogger/v2"

	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix/id"

	"github.com/beeper/groupme-lib"
)

type UserQuery struct {
	db  *Database
	log log.Logger
}

func (uq *UserQuery) New() *User {
	return &User{db: uq.db, log: uq.log}
}

const (
	userColumns        = "gmid, mxid, auth_token, management_room, space_room"
	getAllUsersQuery   = "SELECT " + userColumns + ` FROM "user"`
	getUserByMXIDQuery = getAllUsersQuery + ` WHERE mxid=$1`
	getUserByGMIDQuery = getAllUsersQuery + ` WHERE gmid=$1`
	insertUserQuery    = `INSERT INTO "user" (` + userColumns + `) VALUES ($1, $2, $3, $4, $5)`
	updateUserQurey    = `
		UPDATE "user"
		SET gmid=$1, auth_token=$2, management_room=$3, space_room=$4
		WHERE mxid=$5
	`
)

func (uq *UserQuery) GetAll() (users []*User) {
	rows, err := uq.db.Query(getAllUsersQuery)
	if err != nil || rows == nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		users = append(users, uq.New().Scan(rows))
	}
	return
}

func (uq *UserQuery) GetByMXID(userID id.UserID) *User {
	row := uq.db.QueryRow(getUserByMXIDQuery, userID)
	if row == nil {
		return nil
	}
	return uq.New().Scan(row)
}

func (uq *UserQuery) GetByGMID(gmid groupme.ID) *User {
	row := uq.db.QueryRow(getUserByGMIDQuery, gmid)
	if row == nil {
		return nil
	}
	return uq.New().Scan(row)
}

type User struct {
	db  *Database
	log log.Logger

	MXID           id.UserID
	GMID           groupme.ID
	ManagementRoom id.RoomID
	SpaceRoom      id.RoomID

	Token string

	lastReadCache     map[PortalKey]time.Time
	lastReadCacheLock sync.Mutex
	inSpaceCache      map[PortalKey]bool
	inSpaceCacheLock  sync.Mutex
}

func (user *User) Scan(row dbutil.Scannable) *User {
	var gmid, authToken sql.NullString
	err := row.Scan(&gmid, &user.MXID, &authToken, &user.ManagementRoom, &user.SpaceRoom)
	if err != nil {
		if err != sql.ErrNoRows {
			user.log.Errorln("Database scan failed:", err)
		}
		return nil
	}
	if len(gmid.String) > 0 {
		user.GMID = groupme.ID(gmid.String)
	}
	user.Token = authToken.String
	return user
}

func stripSuffix(gmid groupme.ID) string {
	if len(gmid) == 0 {
		return gmid.String()
	}

	index := strings.IndexRune(gmid.String(), '@')
	if index < 0 {
		return gmid.String()
	}

	return gmid.String()[:index]
}

func (user *User) gmidPtr() *string {
	if len(user.GMID) > 0 {
		str := stripSuffix(user.GMID)
		return &str
	}
	return nil
}

func (user *User) Insert() {
	_, err := user.db.Exec(insertUserQuery, user.gmidPtr(), user.MXID, user.Token, user.ManagementRoom, user.SpaceRoom)
	if err != nil {
		user.log.Warnfln("Failed to insert %s: %v", user.MXID, err)
	}
}

func (user *User) Update() {
	_, err := user.db.Exec(updateUserQurey, user.gmidPtr(), user.Token, user.ManagementRoom, user.SpaceRoom, user.MXID)
	if err != nil {
		user.log.Warnfln("Failed to update %s: %v", user.MXID, err)
	}
}
