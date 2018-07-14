package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// By - тип функции сравнения
type By func(p1, p2 *os.FileInfo) bool

// Sort - метод типа By, который сортирует слайс соответствующей функцией by
func (by By) Sort(fileInfos []os.FileInfo) {
	fileInfoSorter := &fileInfoSorter{
		fileInfos: fileInfos,
		by:        by,
	}

	sort.Sort(fileInfoSorter)
}

type fileInfoSorter struct {
	fileInfos []os.FileInfo
	by        func(p1, p2 *os.FileInfo) bool
}

func (fis *fileInfoSorter) Len() int { return len(fis.fileInfos) }
func (fis *fileInfoSorter) Swap(i, j int) {
	fis.fileInfos[i], fis.fileInfos[j] = fis.fileInfos[j], fis.fileInfos[i]
}
func (fis *fileInfoSorter) Less(i, j int) bool {
	return fis.by(&fis.fileInfos[i], &fis.fileInfos[j])
}

// Элементы для построяния графики
const (
	VerticalRightHalf string = "└"
	VerticalRightFull        = "├"
	Horizonal                = "─"
	Vertical                 = "│"
	Tab                      = "\t"
)

var (
	rightFullGraphics string
	rightHalfGraphics string
)

// TODO: Добавить проверку на ошибки при выводе
var logger func(string) error

func dirTree(out io.Writer, path string, printFiles bool) (resError error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	// Безопасное закрытие файла с помощью defer
	defer func() {
		err := f.Close()
		if resError == nil {
			resError = err
		}
	}()

	fStat, err := f.Stat()
	if err != nil {
		return err
	}

	if !fStat.IsDir() {
		return fmt.Errorf("Can't open file as dir: %s", fStat.Name())
	}

	// Компаратор для сравнения по ключу = name
	fileNameKey := func(p1, p2 *os.FileInfo) bool {
		return (*p1).Name() < (*p2).Name()
	}

	return printDirTree(out, path, "", fileNameKey, printFiles)

}

func printDirTree(out io.Writer, path, mainPrefix string, comparator func(p1, p2 *os.FileInfo) bool,
	printFiles bool) (resError error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		err := f.Close()
		if resError == nil {
			resError = err
		}
	}()

	fStat, err := f.Stat()
	if err != nil {
		return err
	}

	if fStat.IsDir() {
		subFiles, err := f.Readdir(0)
		if err != nil {
			return err
		}

		if len(subFiles) == 0 {
			return nil
		}

		By(comparator).Sort(subFiles)

		if !printFiles {
			// Удаляем файлы из списка
			tmp := make([]os.FileInfo, 0, len(subFiles))
			for _, fInfo := range subFiles {
				if fInfo.IsDir() {
					tmp = append(tmp, fInfo)
				}
			}
			subFiles = tmp[:]
		}

		for idx, fileStat := range subFiles {

			var fileName = fileStat.Name()

			var graphicsPrefix = rightFullGraphics
			var paddingPrefix = Vertical + "\t"

			if idx == len(subFiles)-1 {
				graphicsPrefix = rightHalfGraphics
				paddingPrefix = "\t"
			}

			totalPrefix := mainPrefix + graphicsPrefix

			// Если это файл, то выводим с размером.
			if !fileStat.IsDir() {
				if printFiles {
					formatedFileName := formatFileNameWithSize(fileName, fileStat.Size())
					err = printWithPrefix(out, formatedFileName, totalPrefix)
				} else {
					continue
				}
			} else {
				err = printWithPrefix(out, fileName, totalPrefix)
			}

			// Нет смысла проверять два раза в двух условиях
			if err != nil {
				return err
			}

			err = printDirTree(out, path+string(os.PathSeparator)+fileName, mainPrefix+paddingPrefix, comparator, printFiles)
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func formatFileNameWithSize(fileName string, size int64) string {
	if size == 0 {
		return fmt.Sprintf("%s (empty)", fileName)
	}
	return fmt.Sprintf("%s (%db)", fileName, size)
}

func printWithPrefix(out io.Writer, name, prefix string) error {
	if n, err := fmt.Fprint(out, prefix, name, "\n"); err != nil {
		return err
	} else if n != (len(prefix) + len(name) + 1) {
		return fmt.Errorf("Invalid write \"%s\",\"%s\" to %#v", prefix, name, out)
	}

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func init() {
	buildGraphicsElement(4)

	var prefixer = func(prefix string) func(string) error {
		return func(str string) error {
			if n, err := fmt.Printf("[%s]: %s\n", prefix, str); err != nil {
				return err
			} else if n != len("[]: \n")+len(prefix)+len(str) {
				return fmt.Errorf("Ivalid write \"[%s]: %s\"", prefix, str)
			}
			return nil
		}
	}

	logger = prefixer("LOGGER")
}

// Width ширина графического элемента в (Horizontal + Vertical)
// Width = ( width - 1 ) *Horizontal + Vertical

func buildGraphicsElement(width int) {
	rightFullGraphics = rightFullGraphics + VerticalRightFull + strings.Repeat(Horizonal, width-1)
	rightHalfGraphics = rightHalfGraphics + VerticalRightHalf + strings.Repeat(Horizonal, width-1)
}
