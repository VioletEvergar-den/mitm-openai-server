package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_ensure_dir")
	defer os.RemoveAll(testDir)

	// 测试创建目录
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir失败: %v", err)
	}

	// 验证目录是否存在
	if !DirExists(testDir) {
		t.Errorf("EnsureDir后目录应该存在，但不存在: %s", testDir)
	}

	// 测试重复创建目录（不应该出错）
	err = EnsureDir(testDir)
	if err != nil {
		t.Fatalf("重复调用EnsureDir应该成功，但失败了: %v", err)
	}

	// 测试创建嵌套目录
	nestedDir := filepath.Join(testDir, "nested", "dir")
	err = EnsureDir(nestedDir)
	if err != nil {
		t.Fatalf("创建嵌套目录失败: %v", err)
	}

	if !DirExists(nestedDir) {
		t.Errorf("嵌套目录应该存在，但不存在: %s", nestedDir)
	}
}

func TestFileExists(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_file_exists")
	defer os.RemoveAll(testDir)

	// 确保目录存在
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 测试不存在的文件
	nonExistentFile := filepath.Join(testDir, "non_existent_file.txt")
	if FileExists(nonExistentFile) {
		t.Errorf("不存在的文件应该返回false，但返回了true: %s", nonExistentFile)
	}

	// 创建测试文件
	testFile := filepath.Join(testDir, "test_file.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试存在的文件
	if !FileExists(testFile) {
		t.Errorf("存在的文件应该返回true，但返回了false: %s", testFile)
	}

	// 测试目录（应该返回false，因为不是文件）
	if FileExists(testDir) {
		t.Errorf("目录应该返回false，但返回了true: %s", testDir)
	}
}

func TestDirExists(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_dir_exists")
	defer os.RemoveAll(testDir)

	// 测试不存在的目录
	if DirExists(testDir) {
		t.Errorf("不存在的目录应该返回false，但返回了true: %s", testDir)
	}

	// 创建测试目录
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 测试存在的目录
	if !DirExists(testDir) {
		t.Errorf("存在的目录应该返回true，但返回了false: %s", testDir)
	}

	// 创建测试文件
	testFile := filepath.Join(testDir, "test_file.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试文件（应该返回false，因为不是目录）
	if DirExists(testFile) {
		t.Errorf("文件应该返回false，但返回了true: %s", testFile)
	}
}

func TestCopyFile(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_copy_file")
	defer os.RemoveAll(testDir)

	// 确保目录存在
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 创建源文件
	srcContent := "这是测试内容"
	srcFile := filepath.Join(testDir, "source.txt")
	err = os.WriteFile(srcFile, []byte(srcContent), 0644)
	if err != nil {
		t.Fatalf("创建源文件失败: %v", err)
	}

	// 测试复制到同一目录
	dstFile := filepath.Join(testDir, "destination.txt")
	err = CopyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("复制文件失败: %v", err)
	}

	// 验证目标文件存在
	if !FileExists(dstFile) {
		t.Errorf("复制后目标文件应该存在，但不存在: %s", dstFile)
	}

	// 验证内容正确
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("读取目标文件失败: %v", err)
	}

	if string(dstContent) != srcContent {
		t.Errorf("目标文件内容不正确。期望: %s, 实际: %s", srcContent, string(dstContent))
	}

	// 测试复制到嵌套目录
	nestedDir := filepath.Join(testDir, "nested")
	nestedFile := filepath.Join(nestedDir, "nested_destination.txt")
	err = CopyFile(srcFile, nestedFile)
	if err != nil {
		t.Fatalf("复制到嵌套目录失败: %v", err)
	}

	// 验证嵌套目录中的文件存在
	if !FileExists(nestedFile) {
		t.Errorf("复制后嵌套目录中的文件应该存在，但不存在: %s", nestedFile)
	}
}

func TestListFiles(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_list_files")
	defer os.RemoveAll(testDir)

	// 确保目录存在
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 创建子目录
	subDir := filepath.Join(testDir, "subdir")
	err = EnsureDir(subDir)
	if err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}

	// 创建测试文件
	files := []struct {
		name    string
		content string
	}{
		{"file1.txt", "content1"},
		{"file2.txt", "content2"},
		{"file3.doc", "content3"},
		{"file4.txt", "content4"},
	}

	// 在主目录中创建文件
	for _, f := range files {
		err = os.WriteFile(filepath.Join(testDir, f.name), []byte(f.content), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
	}

	// 在子目录中创建一个文件
	err = os.WriteFile(filepath.Join(subDir, "subfile.txt"), []byte("subcontent"), 0644)
	if err != nil {
		t.Fatalf("在子目录中创建测试文件失败: %v", err)
	}

	// 测试不带模式的列表
	allFiles, err := ListFiles(testDir, "")
	if err != nil {
		t.Fatalf("ListFiles失败: %v", err)
	}

	// 应该有4个文件
	if len(allFiles) != 4 {
		t.Errorf("期望列出4个文件，实际列出了%d个", len(allFiles))
	}

	// 测试带模式的列表（只列出.txt文件）
	txtFiles, err := ListFiles(testDir, "*.txt")
	if err != nil {
		t.Fatalf("ListFiles带模式失败: %v", err)
	}

	// 应该有3个.txt文件
	if len(txtFiles) != 3 {
		t.Errorf("期望列出3个.txt文件，实际列出了%d个", len(txtFiles))
	}
}

func TestGetFileExtension(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"file.txt", "txt"},
		{"/path/to/file.doc", "doc"},
		{"noextension", ""},
		{"/path/with.dots/file", ""},
		{"/path/with.dots/file.js", "js"},
		{".hidden", ""},
		{".hidden.conf", "conf"},
	}

	for _, tc := range testCases {
		result := GetFileExtension(tc.path)
		if result != tc.expected {
			t.Errorf("GetFileExtension(%s): 期望 %s, 实际 %s", tc.path, tc.expected, result)
		}
	}
}

func TestReadWriteStringToFile(t *testing.T) {
	// 创建临时测试目录
	testDir := filepath.Join(os.TempDir(), "test_read_write")
	defer os.RemoveAll(testDir)

	// 确保目录存在
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 测试写入内容
	testFile := filepath.Join(testDir, "test_file.txt")
	testContent := "这是测试内容\n第二行\n第三行"

	err = WriteStringToFile(testFile, testContent)
	if err != nil {
		t.Fatalf("WriteStringToFile失败: %v", err)
	}

	// 验证文件存在
	if !FileExists(testFile) {
		t.Errorf("写入后文件应该存在，但不存在: %s", testFile)
	}

	// 测试读取内容
	readContent, err := ReadFileToString(testFile)
	if err != nil {
		t.Fatalf("ReadFileToString失败: %v", err)
	}

	// 验证内容正确
	if readContent != testContent {
		t.Errorf("读取的内容不正确。期望:\n%s\n实际:\n%s", testContent, readContent)
	}

	// 测试写入到嵌套目录
	nestedFile := filepath.Join(testDir, "nested", "dir", "nested_file.txt")
	err = WriteStringToFile(nestedFile, testContent)
	if err != nil {
		t.Fatalf("写入到嵌套目录失败: %v", err)
	}

	// 验证嵌套目录中的文件存在
	if !FileExists(nestedFile) {
		t.Errorf("写入后嵌套目录中的文件应该存在，但不存在: %s", nestedFile)
	}
}

func TestGetFileName(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"file.txt", "file"},
		{"/path/to/file.doc", "file"},
		{"noextension", "noextension"},
		{"/path/with.dots/file", "file"},
		{"/path/with.dots/file.js", "file"},
		{".hidden", ".hidden"},
		{".hidden.conf", ".hidden"},
	}

	for _, tc := range testCases {
		result := GetFileName(tc.path)
		if result != tc.expected {
			t.Errorf("GetFileName(%s): 期望 %s, 实际 %s", tc.path, tc.expected, result)
		}
	}
}

func TestJoinPaths(t *testing.T) {
	testCases := []struct {
		paths    []string
		expected string
	}{
		{[]string{"path", "to", "file"}, filepath.Join("path", "to", "file")},
		{[]string{"/absolute", "path"}, filepath.Join("/absolute", "path")},
		{[]string{"single"}, "single"},
		{[]string{}, ""},
	}

	for _, tc := range testCases {
		result := JoinPaths(tc.paths...)
		if result != tc.expected {
			t.Errorf("JoinPaths(%v): 期望 %s, 实际 %s", tc.paths, tc.expected, result)
		}
	}
}

func TestIsAbsolutePath(t *testing.T) {
	// 由于不同操作系统的绝对路径格式不同，我们需要分别测试
	if os.PathSeparator == '/' { // Unix-like
		if !IsAbsolutePath("/absolute/path") {
			t.Errorf("'/absolute/path'应该被识别为绝对路径")
		}
		if IsAbsolutePath("relative/path") {
			t.Errorf("'relative/path'不应该被识别为绝对路径")
		}
	} else if os.PathSeparator == '\\' { // Windows
		if !IsAbsolutePath("C:\\absolute\\path") {
			t.Errorf("'C:\\absolute\\path'应该被识别为绝对路径")
		}
		if IsAbsolutePath("relative\\path") {
			t.Errorf("'relative\\path'不应该被识别为绝对路径")
		}
	}
}
