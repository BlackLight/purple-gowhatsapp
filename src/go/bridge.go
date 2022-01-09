/*
 *   gowhatsapp plugin for libpurple
 *   Copyright (C) 2022 Hermann Höhne
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

/*
 * Apparently, the main package must be named main, even though this is a library
 */
package main

/*
#include "constants.h"
#include "bridge.h"

// for feeding messages from go into purple
extern void gowhatsapp_process_message_bridge(gowhatsapp_message_t gwamsg);

// for querying current settings
// these signatures are redefinitions taken from purple.h
// CGO needs to have them re-declared as external
#ifndef _PURPLE_ACCOUNT_H_
struct _PurpleAccount;
extern struct _PurpleAccount * gowhatsapp_get_account(char *username);
extern int purple_account_get_int(const struct _PurpleAccount *account, const char *name, int default_value);
extern char * purple_account_get_string(const struct _PurpleAccount *account, const char *name, char *default_value);
#endif
*/
import "C"

import "time"

// TODO: find out how to enable C99's bool type in cgo
func bool_to_Cchar(b bool) C.char {
	if b {
		return C.char(1)
	} else {
		return C.char(0)
	}
}

func Cint_to_bool(i C.int) bool {
	return i != 0
}

//export gowhatsapp_go_init
func gowhatsapp_go_init(purple_user_dir *C.char) C.int {
	return C.int(init_(C.GoString(purple_user_dir)))
}

//export gowhatsapp_go_login
func gowhatsapp_go_login(username *C.char) {
	login(C.GoString(username))
}

//export gowhatsapp_go_close
func gowhatsapp_go_close(username *C.char) {
	close(C.GoString(username))
}

//export gowhatsapp_go_send_message
func gowhatsapp_go_send_message(username *C.char, who *C.char, message *C.char, is_group C.int) int {
	handler, ok := handlers[C.GoString(username)]
	if ok {
		go handler.send_message(C.GoString(who), C.GoString(message), Cint_to_bool(is_group))
		return 0
	}
	return -107 // ENOTCONN, see libpurple/prpl.h
}

//export gowhatsapp_go_send_file
func gowhatsapp_go_send_file(username *C.char, who *C.char, filename *C.char) int {
	handler, ok := handlers[C.GoString(username)]
	if ok {
		return handler.send_file(C.GoString(who), C.GoString(filename))
	}
	return -107 // ENOTCONN, see libpurple/prpl.h
}

/*
 * This will display a QR code via PurpleRequest API.
 */
func purple_display_qrcode(username string, challenge string, png []byte, terminal string) {
	cmessage := C.struct_gowhatsapp_message{
		username: C.CString(username),
		msgtype:  C.char(C.gowhatsapp_message_type_login),
		text:     C.CString(challenge),
		name:     C.CString(terminal),
		blob:     C.CBytes(png),
		blobsize: C.size_t(len(png)),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will inform purple that the connection has been established.
 */
func purple_connected(username string) {
	cmessage := C.struct_gowhatsapp_message{
		username: C.CString(username),
		msgtype:  C.char(C.gowhatsapp_message_type_connected),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will inform purple that the connection has been destroyed.
 */
func purple_disconnected(username string) {
	cmessage := C.struct_gowhatsapp_message{
		username: C.CString(username),
		msgtype:  C.char(C.gowhatsapp_message_type_disconnected),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will display a text message.
 * Single participants and group chats.
 */
func purple_display_text_message(username string, remoteJid string, isGroup bool, isFromMe bool, senderJid string, pushName *string, timestamp time.Time, text string) {
	cmessage := C.struct_gowhatsapp_message{
		username:  C.CString(username),
		msgtype:   C.char(C.gowhatsapp_message_type_text),
		remoteJid: C.CString(remoteJid),
		senderJid: C.CString(senderJid),
		timestamp: C.time_t(timestamp.Unix()),
		text:      C.CString(text),
		isGroup:   bool_to_Cchar(isGroup),
		fromMe:    bool_to_Cchar(isFromMe),
	}
	if pushName != nil {
		cmessage.name = C.CString(*pushName)
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will display a text message.
 * Single participants and group chats.
 */
func purple_update_name(username string, remoteJid string, pushName string) {
	cmessage := C.struct_gowhatsapp_message{
		username:  C.CString(username),
		msgtype:   C.char(C.gowhatsapp_message_type_name),
		remoteJid: C.CString(remoteJid),
		name:      C.CString(pushName),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will display a text message.
 * Single participants and group chats.
 */
func purple_handle_attachment(username string, senderJid string, filename string, data []byte) {
	cmessage := C.struct_gowhatsapp_message{
		username:  C.CString(username),
		msgtype:   C.char(C.gowhatsapp_message_type_attachment),
		senderJid: C.CString(senderJid),
		name:      C.CString(filename),
		blob:      C.CBytes(data),
		blobsize:  C.size_t(len(data)), // contrary to https://golang.org/pkg/builtin/#len and https://golang.org/ref/spec#Numeric_types, len returns an int of 64 bits on 32 bit Windows machines (see https://github.com/hoehermann/purple-gowhatsapp/issues/1)
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will inform purple that the remote user started typing.
 */
func purple_composing(username string, remoteJid string) {
	cmessage := C.struct_gowhatsapp_message{
		username:  C.CString(username),
		msgtype:   C.char(C.gowhatsapp_message_type_typing),
		remoteJid: C.CString(remoteJid),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will inform purple that the remote user stopped typing.
 */
func purple_paused(username string, remoteJid string) {
	cmessage := C.struct_gowhatsapp_message{
		username:  C.CString(username),
		msgtype:   C.char(C.gowhatsapp_message_type_typing_stopped),
		remoteJid: C.CString(remoteJid),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * This will inform purple that the remote user's presence (online/offline) changed.
 */
func purple_update_presence(username string, remoteJid string, online bool, lastSeen time.Time) {
	cmessage := C.struct_gowhatsapp_message{
		username:  C.CString(username),
		msgtype:   C.char(C.gowhatsapp_message_type_presence),
		remoteJid: C.CString(remoteJid),
		timestamp: C.time_t(lastSeen.Unix()),
		level:     bool_to_Cchar(online),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * Print debug information via purple.
 */
func purple_debug(loglevel int, message string) {
	cmessage := C.struct_gowhatsapp_message{
		msgtype: C.char(C.gowhatsapp_message_type_log),
		level:   C.char(loglevel),
		text:    C.CString(message),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * Forward error to purple. This will cause a disconnect.
 */
func purple_error(username string, message string) {
	cmessage := C.struct_gowhatsapp_message{
		username: C.CString(username),
		msgtype:  C.char(C.gowhatsapp_message_type_error),
		text:     C.CString(message),
	}
	C.gowhatsapp_process_message_bridge(cmessage)
}

/*
 * Get int from the purple account's settings.
 */
func purple_get_int(username string, key *C.char, default_value int) int {
	account := C.gowhatsapp_get_account(C.CString(username))
	return int(C.purple_account_get_int(account, key, C.int(default_value)))
}

/*
 * Get string from the purple account's settings.
 */
func purple_get_string(username string, key *C.char, default_value *C.char) string {
	account := C.gowhatsapp_get_account(C.CString(username))
	return C.GoString(C.purple_account_get_string(account, key, default_value))
}

func main() {
}
