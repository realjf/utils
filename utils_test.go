// #############################################################################
// # File: utils_test.go                                                       #
// # Project: utils                                                            #
// # Created Date: 2024/09/27 10:36:45                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2024/09/27 10:45:51                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// #                                                                           #
// #############################################################################
package utils_test

import (
	"testing"

	"github.com/realjf/utils"
)

func TestGetIPAddress(t *testing.T) {
	addrs, err := utils.GetIPAddress()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("%#v", addrs)
}
