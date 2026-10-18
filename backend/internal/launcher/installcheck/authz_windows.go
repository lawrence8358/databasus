//go:build windows

package installcheck

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	authzResourceManagerNoAudit = 0x1
	maximumAllowed              = 0x02000000
)

var (
	authzLibrary                       = windows.NewLazySystemDLL("authz.dll")
	procAuthzInitializeResourceManager = authzLibrary.NewProc("AuthzInitializeResourceManager")
	procAuthzInitializeContextFromSid  = authzLibrary.NewProc("AuthzInitializeContextFromSid")
	procAuthzAccessCheck               = authzLibrary.NewProc("AuthzAccessCheck")
	procAuthzFreeContext               = authzLibrary.NewProc("AuthzFreeContext")
	procAuthzFreeResourceManager       = authzLibrary.NewProc("AuthzFreeResourceManager")
)

type authzAccessRequest struct {
	desiredAccess        uint32
	principalSelfSid     *windows.SID
	objectTypeList       uintptr
	objectTypeListLength uint32
	optionalArguments    uintptr
}

type authzAccessReply struct {
	resultListLength      uint32
	grantedAccessMask     *uint32
	saclEvaluationResults *uint32
	errorCode             *uint32
}

// getGrantedRights evaluates the file's DACL for an account the way Windows would when that
// account opens it, including the groups the account belongs to.
func getGrantedRights(path string, account *windows.SID) (uint32, error) {
	securityDescriptor, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return 0, err
	}

	var resourceManager uintptr

	result, _, callErr := procAuthzInitializeResourceManager.Call(
		authzResourceManagerNoAudit, 0, 0, 0, 0, uintptr(unsafe.Pointer(&resourceManager)),
	)
	if result == 0 {
		return 0, fmt.Errorf("AuthzInitializeResourceManager: %w", callErr)
	}
	defer func() { _, _, _ = procAuthzFreeResourceManager.Call(resourceManager) }()

	var clientContext uintptr

	result, _, callErr = procAuthzInitializeContextFromSid.Call(
		0, uintptr(unsafe.Pointer(account)), resourceManager, 0, 0, 0, uintptr(unsafe.Pointer(&clientContext)),
	)
	if result == 0 {
		return 0, fmt.Errorf("AuthzInitializeContextFromSid: %w", callErr)
	}
	defer func() { _, _, _ = procAuthzFreeContext.Call(clientContext) }()

	request := authzAccessRequest{desiredAccess: maximumAllowed}

	var grantedAccess, accessError uint32

	reply := authzAccessReply{resultListLength: 1, grantedAccessMask: &grantedAccess, errorCode: &accessError}

	result, _, callErr = procAuthzAccessCheck.Call(
		0,
		clientContext,
		uintptr(unsafe.Pointer(&request)),
		0,
		uintptr(unsafe.Pointer(securityDescriptor)),
		0,
		0,
		uintptr(unsafe.Pointer(&reply)),
		0,
	)
	if result == 0 {
		return 0, fmt.Errorf("AuthzAccessCheck: %w", callErr)
	}

	return grantedAccess, nil
}
