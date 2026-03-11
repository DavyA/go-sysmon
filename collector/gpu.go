package collector

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <string.h>

static void* nvml_handle = NULL;

int load_nvml_c() {
    nvml_handle = dlopen("libnvidia-ml.so", RTLD_LAZY | RTLD_LOCAL);
    if (!nvml_handle) nvml_handle = dlopen("libnvidia-ml.so.1", RTLD_LAZY | RTLD_LOCAL);
    if (!nvml_handle) return -1;

    int (*init_v2)() = dlsym(nvml_handle, "nvmlInit_v2");
    if (!init_v2) {
		int (*init)() = dlsym(nvml_handle, "nvmlInit");
		if(!init) return -2;
		if(init() != 0) return -3;
		return 0;
	}
    if (init_v2() != 0) return -3;
    return 0;
}

int get_nv_gpu_count_c() {
	if(!nvml_handle) return 0;
	int (*get_count)(unsigned int*) = dlsym(nvml_handle, "nvmlDeviceGetCount_v2");
	if(!get_count) return 0;
	unsigned int count = 0;
	get_count(&count);
	return count;
}

typedef struct {
    unsigned int gpu;
    unsigned int memory;
} nvmlUtilization_t;

typedef struct {
    unsigned long long total;
    unsigned long long free;
    unsigned long long used;
} nvmlMemory_t;

int get_nv_gpu_info_c(int index, char* namebuf, int namelen, unsigned int* util, unsigned long long* memtotal, unsigned long long* memused, unsigned int* temp) {
    if(!nvml_handle) return -1;
    int (*get_handle)(unsigned int, void**) = dlsym(nvml_handle, "nvmlDeviceGetHandleByIndex_v2");
    void* dev = NULL;
    if(!get_handle || get_handle(index, &dev) != 0) return -2;

    int (*get_name)(void*, char*, unsigned int) = dlsym(nvml_handle, "nvmlDeviceGetName");
    if(get_name) get_name(dev, namebuf, namelen);

    int (*get_util)(void*, nvmlUtilization_t*) = dlsym(nvml_handle, "nvmlDeviceGetUtilizationRates");
    nvmlUtilization_t u = {0};
    if(get_util) get_util(dev, &u);
    *util = u.gpu;

    int (*get_mem)(void*, nvmlMemory_t*) = dlsym(nvml_handle, "nvmlDeviceGetMemoryInfo");
    nvmlMemory_t m = {0};
    if(get_mem) get_mem(dev, &m);
    *memtotal = m.total;
    *memused = m.used;

    int (*get_temp)(void*, int, unsigned int*) = dlsym(nvml_handle, "nvmlDeviceGetTemperature");
    unsigned int t = 0;
    // 0 is NVML_TEMPERATURE_GPU
    if(get_temp) get_temp(dev, 0, &t);
    *temp = t;

    return 0;
}
*/
import "C"
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type GPUStatus struct {
	ID          string
	Name        string
	Vendor      string
	Driver      string
	Utilization float64 // %
	MemTotal    uint64  // MB
	MemUsed     uint64  // MB
	Temp        float64 // Celsius
	Power       float64 // Watts
}

var nvmlLoaded bool = false
var nvmlAttempted bool = false

func initNVML() {
	if nvmlAttempted {
		return
	}
	res := C.load_nvml_c()
	if res == 0 {
		nvmlLoaded = true
	}
	nvmlAttempted = true
}

func GetAllGPUs() []GPUStatus {
	var gpus []GPUStatus

	// Precarichiamo i device PCI per identificare il nome preciso della scheda open-source
	pciDevs, _ := GetGraphicsDevices()

	// 1. Try NVML via dlopen (closed source NVIDIA)
	initNVML()
	if nvmlLoaded {
		count := int(C.get_nv_gpu_count_c())
		for i := 0; i < count; i++ {
			namebuf := make([]byte, 128)
			var util C.uint
			var memTot, memUsd C.ulonglong
			var temp C.uint

			res := C.get_nv_gpu_info_c(C.int(i), (*C.char)(unsafe.Pointer(&namebuf[0])), 128, &util, &memTot, &memUsd, &temp)
			if res == 0 {
				name := C.GoString((*C.char)(unsafe.Pointer(&namebuf[0])))

				// TODO: Fetch power consumption per NVIDIA NVML (nvmlDeviceGetPowerUsage) - limitato qui dal C
				gpus = append(gpus, GPUStatus{
					ID:          fmt.Sprintf("nv-gpu%d", i),
					Name:        name,
					Driver:      "nvidia (nvml)",
					Vendor:      "NVIDIA",
					Utilization: float64(util),
					MemTotal:    uint64(memTot) / 1024 / 1024,
					MemUsed:     uint64(memUsd) / 1024 / 1024,
					Temp:        float64(temp),
					Power:       0.0, // Non ancora mappato dal C
				})
			}
		}
	}

	// 2. Fallback / Addition: Sysfs / DRM for open source drivers (Nouveau, AMD, Intel)
	cards, _ := filepath.Glob("/sys/class/drm/card[0-9]*")
	for _, cardPath := range cards {
		devPath := filepath.Join(cardPath, "device")

		vendorBytes, err := os.ReadFile(filepath.Join(devPath, "vendor"))
		if err != nil {
			continue
		}
		vendorCode := strings.TrimSpace(string(vendorBytes))

		driverLink, err := os.Readlink(filepath.Join(devPath, "driver"))
		driver := "unknown"
		if err == nil {
			driver = filepath.Base(driverLink)
		}

		if nvmlLoaded && strings.Contains(strings.ToLower(driver), "nvidia") {
			continue // Già catturata via nvml
		}

		vendor := "Unknown"
		switch vendorCode {
		case "0x1002":
			vendor = "AMD"
		case "0x10de":
			vendor = "NVIDIA"
		case "0x8086":
			vendor = "Intel"
		}

		// Map back to specific PCI model name
		cardName := fmt.Sprintf("%s GPU (%s)", vendor, driver)
		realPath, err := filepath.EvalSymlinks(devPath)
		if err == nil {
			// realPath potrebbe essere .../0000:c3:00.0
			slotPart := filepath.Base(realPath) // "0000:c3:00.0"
			// su lspci è "c3:00.0" oppure "0000:c3:00.0"
			for _, pci := range pciDevs {
				if strings.Contains(slotPart, pci.Address) {
					cardName = pci.Name
					break
				}
			}
		}

		var utilization float64
		busyBytes, err := os.ReadFile(filepath.Join(devPath, "gpu_busy_percent"))
		if err == nil {
			fmt.Sscanf(string(busyBytes), "%f", &utilization)
		}

		memTot := uint64(0)
		memUsd := uint64(0)
		if vendor == "AMD" || vendorCode == "0x1002" {
			vramTot, _ := os.ReadFile(filepath.Join(devPath, "mem_info_vram_total"))
			fmt.Sscanf(string(vramTot), "%d", &memTot)
			vramUsd, _ := os.ReadFile(filepath.Join(devPath, "mem_info_vram_used"))
			fmt.Sscanf(string(vramUsd), "%d", &memUsd)
			memTot = memTot / 1024 / 1024
			memUsd = memUsd / 1024 / 1024
		}

		// Trova le info hwmon per Temperatura e Power (Watt)
		var temp, power float64
		hwmonPath, _ := filepath.Glob(filepath.Join(devPath, "hwmon/hwmon*"))
		if len(hwmonPath) > 0 {
			hPath := hwmonPath[0]

			// Temperatura da temp1_input (in milligradi)
			tBytes, err := os.ReadFile(filepath.Join(hPath, "temp1_input"))
			if err == nil {
				var millicels float64
				fmt.Sscanf(strings.TrimSpace(string(tBytes)), "%f", &millicels)
				temp = millicels / 1000.0
			}

			// Potenza (Watt) in microwatt (power1_average o power1_input)
			pBytes, err := os.ReadFile(filepath.Join(hPath, "power1_average"))
			if err != nil {
				pBytes, _ = os.ReadFile(filepath.Join(hPath, "power1_input"))
			}
			if len(pBytes) > 0 {
				var microwatts float64
				fmt.Sscanf(strings.TrimSpace(string(pBytes)), "%f", &microwatts)
				power = microwatts / 1000000.0
			}
		}

		gpus = append(gpus, GPUStatus{
			ID:          filepath.Base(cardPath),
			Name:        cardName,
			Vendor:      vendor,
			Driver:      driver,
			Utilization: utilization,
			MemTotal:    memTot,
			MemUsed:     memUsd,
			Temp:        temp,
			Power:       power,
		})
	}

	return gpus
}

func GetGpuPIDs() map[int]bool {
	pids := make(map[int]bool)

	// DRM Clients (Mostly AMD/Intel on newer kernels)
	clients, _ := filepath.Glob("/sys/class/drm/card*/clients/*/pid")
	for _, c := range clients {
		data, err := os.ReadFile(c)
		if err == nil {
			pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
			if pid > 0 { pids[pid] = true }
		}
	}

	// NVIDIA specific if NVML not available or just proxy via lsof-style file scan
	// We only do this check for some active processes to keep it fast, or if sysfs failed.
	return pids
}

// GetGPUState background ticker
func GetGPUState(ctx context.Context, ticker <-chan time.Time) chan []GPUStatus {
	output := make(chan []GPUStatus)
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				output <- GetAllGPUs()
			}
		}
	}()
	return output
}
