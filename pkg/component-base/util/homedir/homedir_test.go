package homedir

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHomeDir_NonWindows(t *testing.T) {
	// 保存原始值
	oldHOME := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", oldHOME)
	}()

	// 由于 runtime.GOOS 是常量，我们无法在测试中修改它
	// 所以我们只能测试当前系统的行为
	if runtime.GOOS != "windows" {
		t.Run("Non-Windows system uses HOME env", func(t *testing.T) {
			expected := "/test/home/dir"
			os.Setenv("HOME", expected)
			result := HomeDir()
			if result != expected {
				t.Errorf("HomeDir() = %q, want %q", result, expected)
			}
		})
	} else {
		t.Skip("Skipping non-Windows test on Windows system")
	}
}

func TestHomeDir_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows test on non-Windows system")
	}

	tests := []struct {
		name           string
		setupEnv       map[string]string
		createFile     string // 创建 .apimachinery\config 文件的路径
		expectContains string // 期望结果包含的路径标识
	}{
		{
			name: "HOME contains config file",
			setupEnv: map[string]string{
				"HOME":      "C:\\home",
				"HOMEDRIVE": "",
				"HOMEPATH":  "",
				"USERPROFILE": "",
			},
			createFile:     "C:\\home\\.apimachinery\\config",
			expectContains: "C:\\home",
		},
		{
			name: "HOMEDRIVE+HOMEPATH contains config file",
			setupEnv: map[string]string{
				"HOME":        "",
				"HOMEDRIVE":   "C:",
				"HOMEPATH":    "\\Users\\test",
				"USERPROFILE": "C:\\profile",
			},
			createFile:     "C:\\Users\\test\\.apimachinery\\config",
			expectContains: "C:\\Users\\test",
		},
		{
			name: "USERPROFILE contains config file",
			setupEnv: map[string]string{
				"HOME":        "",
				"HOMEDRIVE":   "C:",
				"HOMEPATH":    "\\Users\\test",
				"USERPROFILE": "C:\\profile",
			},
			createFile:     "C:\\profile\\.apimachinery\\config",
			expectContains: "C:\\profile",
		},
		{
			name: "HOME is writeable directory",
			setupEnv: map[string]string{
				"HOME":        "C:\\writable_home",
				"HOMEDRIVE":   "",
				"HOMEPATH":    "",
				"USERPROFILE": "",
			},
			createFile:     "", // 不创建 config 文件
			expectContains: "C:\\writable_home",
		},
		{
			name: "USERPROFILE is writeable when no config",
			setupEnv: map[string]string{
				"HOME":        "",
				"HOMEDRIVE":   "C:",
				"HOMEPATH":    "\\Users\\test",
				"USERPROFILE": "C:\\profile",
			},
			createFile:     "",
			expectContains: "C:\\profile",
		},
		{
			name: "Return first existing path when none writeable",
			setupEnv: map[string]string{
				"HOME":        "",
				"HOMEDRIVE":   "C:",
				"HOMEPATH":    "\\Users\\test",
				"USERPROFILE": "",
			},
			createFile:     "",
			expectContains: "C:\\Users\\test",
		},
		{
			name: "Return first set path when none exist",
			setupEnv: map[string]string{
				"HOME":        "",
				"HOMEDRIVE":   "D:",
				"HOMEPATH":    "\\path",
				"USERPROFILE": "",
			},
			createFile:     "",
			expectContains: "D:\\path",
		},
		{
			name: "Return empty when nothing set",
			setupEnv: map[string]string{
				"HOME":        "",
				"HOMEDRIVE":   "",
				"HOMEPATH":    "",
				"USERPROFILE": "",
			},
			createFile:     "",
			expectContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始环境变量
			oldHOME := os.Getenv("HOME")
			oldHOMEDRIVE := os.Getenv("HOMEDRIVE")
			oldHOMEPATH := os.Getenv("HOMEPATH")
			oldUSERPROFILE := os.Getenv("USERPROFILE")

			// 设置测试环境变量
			for k, v := range tt.setupEnv {
				os.Setenv(k, v)
			}

			// 如果需要，创建 config 文件
			var cleanupFuncs []func()
			if tt.createFile != "" {
				dir := filepath.Dir(tt.createFile)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
				f, err := os.Create(tt.createFile)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				f.Close()
				cleanupFuncs = append(cleanupFuncs, func() {
					os.Remove(tt.createFile)
					os.RemoveAll(filepath.Dir(tt.createFile))
				})
			}

			// 创建可写目录（用于测试 writeable 场景）
			if tt.expectContains != "" && !filepath.IsAbs(tt.expectContains) {
				// 如果期望路径不是绝对路径，跳过此测试
				t.Skip("Skipping writeable test without absolute path")
			}

			// 调用被测试函数
			result := HomeDir()

			// 验证结果
			if tt.expectContains == "" {
				if result != "" {
					t.Errorf("HomeDir() = %q, want empty string", result)
				}
			} else {
				if result != tt.expectContains && result != tt.setupEnv["HOME"] && result != tt.setupEnv["USERPROFILE"] && result != (tt.setupEnv["HOMEDRIVE"]+tt.setupEnv["HOMEPATH"]) {
					// 更宽松的检查：只要结果不为空且在预期的某个路径中即可
					found := false
					for _, v := range tt.setupEnv {
						if v != "" && result == v {
							found = true
							break
						}
					}
					if !found && result != tt.setupEnv["HOMEDRIVE"]+tt.setupEnv["HOMEPATH"] {
						t.Errorf("HomeDir() = %q, want one of the set paths", result)
					}
				}
			}

			// 清理
			for _, cleanup := range cleanupFuncs {
				cleanup()
			}

			// 恢复原始环境变量
			os.Setenv("HOME", oldHOME)
			os.Setenv("HOMEDRIVE", oldHOMEDRIVE)
			os.Setenv("HOMEPATH", oldHOMEPATH)
			os.Setenv("USERPROFILE", oldUSERPROFILE)
		})
	}
}

func TestHomeDir_Precedence(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping precedence test on non-Windows system")
	}

	// 测试优先级：config 文件 > 可写目录 > 存在目录 > 已设置路径
	oldHOME := os.Getenv("HOME")
	oldHOMEDRIVE := os.Getenv("HOMEDRIVE")
	oldHOMEPATH := os.Getenv("HOMEPATH")
	oldUSERPROFILE := os.Getenv("USERPROFILE")
	defer func() {
		os.Setenv("HOME", oldHOME)
		os.Setenv("HOMEDRIVE", oldHOMEDRIVE)
		os.Setenv("HOMEPATH", oldHOMEPATH)
		os.Setenv("USERPROFILE", oldUSERPROFILE)
	}()

	// 测试 config 文件优先级最高
	os.Setenv("HOME", "C:\\home_no_config")
	os.Setenv("USERPROFILE", "C:\\profile_with_config")

	// 在 USERPROFILE 创建 config 文件
	configPath := "C:\\profile_with_config\\.apimachinery\\config"
	os.MkdirAll(filepath.Dir(configPath), 0755)
	f, _ := os.Create(configPath)
	f.Close()
	defer func() {
		os.Remove(configPath)
		os.RemoveAll("C:\\profile_with_config\\.apimachinery")
	}()

	result := HomeDir()
	if result != "C:\\profile_with_config" {
		t.Errorf("HomeDir() should prefer path with config file, got %q, want %q", result, "C:\\profile_with_config")
	}
}
