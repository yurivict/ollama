package discover

/*
#include <stdlib.h>
#include <sys/types.h>
#include <sys/sysctl.h>
*/
import "C"

import (
	"log/slog"
	"strings"
	"syscall"
	"unsafe"
)

func sysctlUint64(name string) (uint64, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	
	var value C.uint64_t
	size := C.size_t(unsafe.Sizeof(value))
	
	ret := C.sysctlbyname(cname, unsafe.Pointer(&value), &size, nil, 0)
	if ret != 0 {
		return 0, syscall.Errno(ret)
	}
	
	return uint64(value), nil
}

func GetCPUMem() (memInfo, error) {
	var mem memInfo

	// Get page size - this is a 32-bit value
	pageSize32, err := syscall.SysctlUint32("vm.stats.vm.v_page_size")
	if err != nil {
		return mem, err
	}
	pageSize := uint64(pageSize32)

	// Get physical memory - use sysctlUint64
	physmem, err := sysctlUint64("hw.physmem")
	if err != nil {
		return mem, err
	}

	// Get free page count - this is also a 32-bit value
	freeCount32, err := syscall.SysctlUint32("vm.stats.vm.v_free_count")
	if err != nil {
		return mem, err
	}
	freeCount := uint64(freeCount32)

	// Get swap total - use sysctlUint64
	swapTotal, err := sysctlUint64("vm.swap_total")
	if err != nil {
		// Swap may not be configured, default to 0
		swapTotal = 0
	}

	mem.TotalMemory = physmem
	mem.FreeMemory = freeCount * pageSize
	mem.FreeSwap = swapTotal

	slog.Debug("GetCPUMem", "total_memory", mem.TotalMemory, "free_memory", mem.FreeMemory, "free_swap", mem.FreeSwap)

	return mem, nil
}

func GetCPUDetails() []CPU {
	var cpus []CPU

	// Get CPU model name - this is a string
	modelName, err := syscall.Sysctl("hw.model")
	if err != nil {
		slog.Warn("failed to get CPU model", "error", err)
		modelName = "Unknown"
	}

	// Get number of physical cores - this is a 32-bit integer
	cores32, err := syscall.SysctlUint32("kern.smp.cores")
	if err != nil {
		slog.Warn("failed to get CPU cores", "error", err)
		return nil
	}
	cores := int(cores32)

	// Get number of logical CPUs (threads) - this is a 32-bit integer
	threads32, err := syscall.SysctlUint32("hw.ncpu")
	if err != nil {
		slog.Warn("failed to get CPU threads", "error", err)
		return nil
	}
	threads := int(threads32)

	// Extract vendor ID from model name if possible
	vendorID := ""
	modelNameLower := strings.ToLower(modelName)
	if strings.Contains(modelNameLower, "intel") {
		vendorID = "GenuineIntel"
	} else if strings.Contains(modelNameLower, "amd") {
		vendorID = "AuthenticAMD"
	}

	// For FreeBSD, we assume a single socket for now
	// In the future, this could be enhanced to detect multi-socket systems
	cpu := CPU{
		ID:                  "0",
		VendorID:            vendorID,
		ModelName:           strings.TrimSpace(modelName),
		CoreCount:           cores,
		EfficiencyCoreCount: 0, // FreeBSD doesn't distinguish efficiency cores
		ThreadCount:         threads,
	}

	cpus = append(cpus, cpu)

	slog.Debug("GetCPUDetails", "cpus", cpus)

	return cpus
}
