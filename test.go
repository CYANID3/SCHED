package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Task struct {
	Name, Status, LastResult string
}

// Получаем задачи конкретной папки
func getTasks(folder string) []Task {
	cmd := exec.Command("schtasks.exe", "/Query", "/FO", "LIST", "/V")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(string(out), "\n")

	var tasks []Task
	var t Task
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if t.Name != "" {
				// Сравниваем, чтобы оставить только задачи из нужной папки
				if strings.HasPrefix(t.Name, folder) {
					tasks = append(tasks, t)
				}
				t = Task{}
			}
			continue
		}
		if strings.HasPrefix(line, "TaskName:") {
			t.Name = strings.TrimSpace(strings.TrimPrefix(line, "TaskName:"))
		} else if strings.HasPrefix(line, "Status:") {
			t.Status = strings.TrimSpace(strings.TrimPrefix(line, "Status:"))
		} else if strings.HasPrefix(line, "Last Result:") {
			t.LastResult = strings.TrimSpace(strings.TrimPrefix(line, "Last Result:"))
		}
	}
	if t.Name != "" && strings.HasPrefix(t.Name, folder) {
		tasks = append(tasks, t)
	}
	return tasks
}

func main() {
	app := tview.NewApplication()

	root := tview.NewTreeNode("Win Scheduler").SetExpanded(true)
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	tree.SetBorder(true).SetTitle("Папки")

	// ===== Получаем все задачи и строим дерево папок =====
	cmd := exec.Command("schtasks.exe", "/Query", "/FO", "LIST", "/V")
	out, _ := cmd.Output()
	lines := strings.Split(string(out), "\n")

	nodes := map[string]*tview.TreeNode{"": root} // ключ = полный путь
	folderPaths := map[*tview.TreeNode]string{root: ""}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TaskName:") {
			taskName := strings.TrimSpace(strings.TrimPrefix(line, "TaskName:"))
			parts := strings.Split(taskName, "\\")[1:] // убрать пустой корень
			path := ""
			for i, p := range parts[:len(parts)-1] {
				if path == "" {
					path = p
				} else {
					path = path + "\\" + p
				}
				if _, ok := nodes[path]; !ok {
					node := tview.NewTreeNode(p).SetExpanded(true)
					nodes[path] = node
					parentPath := strings.Join(parts[:i], "\\")
					parentNode := nodes[parentPath]
					parentNode.AddChild(node)
					folderPaths[node] = "\\" + path
				}
			}
		}
	}

	table := tview.NewTable().SetBorders(true).SetSelectable(true, false)
	table.SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite))

	var currentFolder string
	var tasks []Task

	refreshTable := func(folder string) {
		currentFolder = folder
		table.Clear()
		table.SetCell(0, 0, tview.NewTableCell("Имя").SetAttributes(tcell.AttrBold).SetTextColor(tcell.ColorWhite))
		table.SetCell(0, 1, tview.NewTableCell("Статус").SetAttributes(tcell.AttrBold).SetTextColor(tcell.ColorWhite))
		table.SetCell(0, 2, tview.NewTableCell("Last Result").SetAttributes(tcell.AttrBold).SetTextColor(tcell.ColorWhite))

		tasks = getTasks(folder)
		if tasks == nil {
			tasks = []Task{}
		}
		for i, t := range tasks {
			cellName := tview.NewTableCell(t.Name)
			cellStatus := tview.NewTableCell(t.Status)
			cellResult := tview.NewTableCell(t.LastResult)

			switch t.Status {
			case "Ready":
				cellStatus.SetTextColor(tcell.ColorGreen)
			case "Running":
				cellStatus.SetTextColor(tcell.ColorYellow)
			case "Disabled":
				cellStatus.SetTextColor(tcell.ColorRed)
			default:
				cellStatus.SetTextColor(tcell.ColorWhite)
			}

			table.SetCell(i+1, 0, cellName)
			table.SetCell(i+1, 1, cellStatus)
			table.SetCell(i+1, 2, cellResult)
		}
	}

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		path, ok := folderPaths[node]
		if !ok || path == "" {
			return
		}
		refreshTable(path)
		app.SetFocus(table)
	})

	// ===== Кнопки как таблица =====
	buttonLabels := []string{"Run All", "Stop All", "Enable All", "Disable All"}
	buttonsTable := tview.NewTable().SetSelectable(true, false)
	buttonsTable.SetBorder(true).SetTitle("Actions")
	for i, label := range buttonLabels {
		buttonsTable.SetCell(0, i, tview.NewTableCell(label).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignCenter))
	}

	activeButton := 0
	updateButtonFocus := func(active int) {
		for i := 0; i < len(buttonLabels); i++ {
			cell := buttonsTable.GetCell(0, i)
			if i == active {
				cell.SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold)
			} else {
				cell.SetTextColor(tcell.ColorWhite).SetAttributes(tcell.AttrNone)
			}
		}
	}

	buttonAction := func(i int) {
		for _, t := range tasks {
			switch i {
			case 0:
				exec.Command("schtasks.exe", "/Run", "/TN", t.Name).Run()
			case 1:
				exec.Command("schtasks.exe", "/End", "/TN", t.Name).Run()
			case 2:
				exec.Command("schtasks.exe", "/Change", "/TN", t.Name, "/ENABLE").Run()
			case 3:
				exec.Command("schtasks.exe", "/Change", "/TN", t.Name, "/DISABLE").Run()
			}
		}
		refreshTable(currentFolder)
	}
	updateButtonFocus(activeButton)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().AddItem(tree, 30, 1, true).AddItem(table, 0, 3, false), 0, 1, true).
		AddItem(buttonsTable, 3, 0, false)

	// Enter на задаче
	table.SetSelectedFunc(func(row, col int) {
		if row == 0 {
			return
		}
		t := tasks[row-1]
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Task: %s\nStatus: %s\nLast Result: %s", t.Name, t.Status, t.LastResult)).
			AddButtons([]string{"Run", "Stop", "Enable", "Disable", "Cancel"}).
			SetDoneFunc(func(index int, label string) {
				switch label {
				case "Run":
					exec.Command("schtasks.exe", "/Run", "/TN", t.Name).Run()
				case "Stop":
					exec.Command("schtasks.exe", "/End", "/TN", t.Name).Run()
				case "Enable":
					exec.Command("schtasks.exe", "/Change", "/TN", t.Name, "/ENABLE").Run()
				case "Disable":
					exec.Command("schtasks.exe", "/Change", "/TN", t.Name, "/DISABLE").Run()
				}
				refreshTable(currentFolder)
				app.SetFocus(table)
				app.SetRoot(mainFlex, true)
			})
		app.SetRoot(modal, true)
	})

	// Управление клавиатурой
	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		focus := app.GetFocus()
		switch ev.Key() {
		case tcell.KeyEsc:
			app.SetFocus(tree) // Esc возвращает на древо
		case tcell.KeyTab:
			if focus == tree {
				app.SetFocus(table)
			} else if focus == table {
				app.SetFocus(buttonsTable)
				activeButton = 0
				updateButtonFocus(activeButton)
			} else {
				app.SetFocus(tree)
			}
		case tcell.KeyRight:
			if focus == buttonsTable && activeButton < len(buttonLabels)-1 {
				activeButton++
				updateButtonFocus(activeButton)
			}
		case tcell.KeyLeft:
			if focus == buttonsTable && activeButton > 0 {
				activeButton--
				updateButtonFocus(activeButton)
			}
		case tcell.KeyEnter:
			if focus == buttonsTable {
				buttonAction(activeButton)
			}
		}
		return ev
	})

	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
		panic(err)
	}
}
