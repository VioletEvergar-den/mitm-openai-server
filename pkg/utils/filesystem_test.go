package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirectory(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_ensure_directory")
	defer os.RemoveAll(testDir)

	// 测试创建目录
	err := EnsureDirectory(testDir)
	if err != nil {
		t.Fatalf("EnsureDirectory失败: %v", err)
	}

	// 验证目录是否存在
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("无法获取目录信息: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("EnsureDirectory后路径应该是目录，但不是: %s", testDir)
	}

	// 测试重复创建目录（不应该出错）
	err = EnsureDirectory(testDir)
	if err != nil {
		t.Fatalf("重复调用EnsureDirectory应该成功，但失败了: %v", err)
	}

	// 测试创建嵌套目录
	nestedDir := filepath.Join(testDir, "nested", "dir")
	err = EnsureDirectory(nestedDir)
	if err != nil {
		t.Fatalf("创建嵌套目录失败: %v", err)
	}

	info, err = os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("无法获取嵌套目录信息: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("嵌套目录应该是目录，但不是: %s", nestedDir)
	}
}

func TestBuildFilePath(t *testing.T) {
	tests := []struct {
		name     string
		baseDir  string
		id       string
		ext      string
		expected string
	}{
		{
			name:     "扩展名带点号",
			baseDir:  "/tmp",
			id:       "test",
			ext:      ".txt",
			expected: filepath.Join("/tmp", "test.txt"),
		},
		{
			name:     "扩展名不带点号",
			baseDir:  "/tmp",
			id:       "test",
			ext:      "txt",
			expected: filepath.Join("/tmp", "test.txt"),
		},
		{
			name:     "空扩展名",
			baseDir:  "/tmp",
			id:       "test",
			ext:      "",
			expected: filepath.Join("/tmp", "test."),
		},
		{
			name:     "相对路径",
			baseDir:  "tmp",
			id:       "test",
			ext:      ".txt",
			expected: filepath.Join("tmp", "test.txt"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildFilePath(tc.baseDir, tc.id, tc.ext)
			if result != tc.expected {
				t.Errorf("BuildFilePath(%q, %q, %q) = %q, 期望 %q", tc.baseDir, tc.id, tc.ext, result, tc.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "安全文件名",
			filename: "test",
			expected: "test",
		},
		{
			name:     "含有路径分隔符",
			filename: "test/file",
			expected: "test_file",
		},
		{
			name:     "含有Windows路径分隔符",
			filename: "test\\file",
			expected: "test_file",
		},
		{
			name:     "含有其他不安全字符",
			filename: "test:*?\"<>|file",
			expected: "test_______file",
		},
		{
			name:     "空文件名",
			filename: "",
			expected: "unnamed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeFilename(tc.filename)
			if result != tc.expected {
				t.Errorf("SanitizeFilename(%q) = %q, 期望 %q", tc.filename, result, tc.expected)
			}
		})
	}
}

func TestIsFileExists(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_is_file_exists")
	defer os.RemoveAll(testDir)

	// 确保目录存在
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 测试不存在的文件
	nonExistentFile := filepath.Join(testDir, "non_existent_file.txt")
	exists, err := IsFileExists(nonExistentFile)
	if err != nil {
		t.Fatalf("IsFileExists对不存在的文件检查失败: %v", err)
	}
	if exists {
		t.Errorf("不存在的文件应该返回false，但返回了true: %s", nonExistentFile)
	}

	// 创建测试文件
	testFile := filepath.Join(testDir, "test_file.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试存在的文件
	exists, err = IsFileExists(testFile)
	if err != nil {
		t.Fatalf("IsFileExists对存在的文件检查失败: %v", err)
	}
	if !exists {
		t.Errorf("存在的文件应该返回true，但返回了false: %s", testFile)
	}

	// 测试目录
	dirExists, err := IsFileExists(testDir)
	if err != nil {
		t.Fatalf("IsFileExists对目录检查失败: %v", err)
	}
	if dirExists {
		t.Errorf("目录路径应该返回false (因为不是文件)，但返回了true: %s", testDir)
	}
}

func TestGetFileSize(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_get_file_size")
	defer os.RemoveAll(testDir)

	// 确保目录存在
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 测试获取不存在文件的大小
	nonExistentFile := filepath.Join(testDir, "non_existent_file.txt")
	_, err := GetFileSize(nonExistentFile)
	if err == nil {
		t.Errorf("GetFileSize应该对不存在的文件返回错误，但未返回: %s", nonExistentFile)
	}

	// 创建测试文件
	testContent := "test content"
	testFile := filepath.Join(testDir, "test_file.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试获取存在文件的大小
	size, err := GetFileSize(testFile)
	if err != nil {
		t.Fatalf("GetFileSize对存在的文件检查失败: %v", err)
	}
	if size != int64(len(testContent)) {
		t.Errorf("文件大小不匹配，得到 %d, 期望 %d", size, len(testContent))
	}
}

func TestCreateTempFile(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_create_temp_file")
	defer os.RemoveAll(testDir)

	// 测试在不存在的目录中创建临时文件
	path, err := CreateTempFile(testDir, "test-")
	if err != nil {
		t.Fatalf("在新目录中CreateTempFile失败: %v", err)
	}

	// 验证文件是否存在
	exists, err := IsFileExists(path)
	if err != nil {
		t.Fatalf("检查创建的临时文件失败: %v", err)
	}
	if !exists {
		t.Errorf("创建的临时文件应该存在，但不存在: %s", path)
	}

	// 检查目录是否被创建
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("无法获取目录信息: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("目录应该被创建，但不是目录: %s", testDir)
	}

	// 测试在系统临时目录中创建临时文件
	path, err = CreateTempFile("", "test-")
	if err != nil {
		t.Fatalf("在系统临时目录中CreateTempFile失败: %v", err)
	}

	// 验证文件是否存在
	exists, err = IsFileExists(path)
	if err != nil {
		t.Fatalf("检查创建的临时文件失败: %v", err)
	}
	if !exists {
		t.Errorf("创建的临时文件应该存在，但不存在: %s", path)
	}
}
