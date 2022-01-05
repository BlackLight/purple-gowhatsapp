/*
 *   gowhatsapp plugin for libpurple
 *   Copyright (C) 2019 Hermann Höhne
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

package main

/*
#include <stdint.h>
*/
import "C"

func login(connID C.uintptr_t) {
	// TODO: protect against concurrent invocation
	handler := waHandler{
		wac:                nil,
		connID:             connID,
	}
	waHandlers[connID] = &handler
    // TODO
}

func close(connID C.uintptr_t) {
	handler, ok := waHandlers[connID]
	if ok {
		if handler.wac != nil {
			handler.wac.Disconnect()
		}
		delete(waHandlers, connID)
	}
}
