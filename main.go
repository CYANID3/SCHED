// package main

// import (
// 	"fmt"

// 	"github.com/gdamore/tcell/v2"
// 	"github.com/rivo/tview"
// )

// func main() {
// 	app := tview.NewApplication()

// 	// ===== Структура папок Scheduler =====
// 	folders := map[string][]string{
// 		"Windows": {"Update", "Defrag"},
// 		"Custom":  {"Backup", "Cleanup"},
// 	}

// 	// ===== Пример задач для каждой папки =====
// 	tasksByFolder := map[string][]struct {
// 		Name, Status, Result string
// 	}{}

// 	for _, folderList := range folders {
// 		for _, f := range folderList {
// 			var tasks []struct {
// 				Name, Status, Result string
// 			}
// 			for i := 1; i <= 10; i++ {
// 				status := "Ready"
// 				if i%3 == 0 {
// 					status = "Running"
// 				} else if i%5 == 0 {
// 					status = "Disabled"
// 				}
// 				tasks = append(tasks, struct {
// 					Name, Status, Result string
// 				}{fmt.Sprintf("%s_Task_%d", f, i), status, fmt.Sprintf("0x%x", i)})
// 			}
// 			tasksByFolder[f] = tasks
// 		}
// 	}

// 	// ===== Дерево =====
// 	root := tview.NewTreeNode("Win Scheduler").SetExpanded(true)
// 	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
// 	tree.SetBorder(true).SetTitle("Папки")

// 	for parent, children := range folders {
// 		parentNode := tview.NewTreeNode(parent).SetExpanded(true)
// 		for _, child := range children {
// 			childNode := tview.NewTreeNode(child)
// 			parentNode.AddChild(childNode)
// 		}
// 		root.AddChild(parentNode)
// 	}

// 	// ===== Таблица задач =====
// 	table := tview.NewTable().SetBorders(true).SetSelectable(true, false)
// 	table.SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite))
// 	var currentFolder string

// 	refreshTable := func(folder string) {
// 		currentFolder = folder
// 		table.Clear()
// 		table.SetCell(0, 0, tview.NewTableCell("Имя").SetAttributes(tcell.AttrBold).SetTextColor(tcell.ColorWhite))
// 		table.SetCell(0, 1, tview.NewTableCell("Статус").SetAttributes(tcell.AttrBold).SetTextColor(tcell.ColorWhite))
// 		table.SetCell(0, 2, tview.NewTableCell("Last Result").SetAttributes(tcell.AttrBold).SetTextColor(tcell.ColorWhite))

// 		tasks := tasksByFolder[folder]
// 		for i, t := range tasks {
// 			cellName := tview.NewTableCell(t.Name)
// 			cellStatus := tview.NewTableCell(t.Status)
// 			cellResult := tview.NewTableCell(t.Result)

// 			switch t.Status {
// 			case "Ready":
// 				cellStatus.SetTextColor(tcell.ColorGreen)
// 			case "Running":
// 				cellStatus.SetTextColor(tcell.ColorYellow)
// 			case "Disabled":
// 				cellStatus.SetTextColor(tcell.ColorRed)
// 			}

// 			table.SetCell(i+1, 0, cellName)
// 			table.SetCell(i+1, 1, cellStatus)
// 			table.SetCell(i+1, 2, cellResult)
// 		}
// 	}

// 	tree.SetSelectedFunc(func(node *tview.TreeNode) {
// 		folder := node.GetText()
// 		if folder != "Win Scheduler" && folder != "Windows" && folder != "Custom" {
// 			refreshTable(folder)
// 			app.SetFocus(table)
// 		}
// 	})

// 	// ===== Кнопки как таблица =====
// 	buttonLabels := []string{"Run All", "Stop All", "Enable All", "Disable All"}
// 	buttonsTable := tview.NewTable().SetSelectable(true, false)
// 	buttonsTable.SetBorder(true).SetTitle("Actions")
// 	for i, label := range buttonLabels {
// 		btn := tview.NewTableCell(label).
// 			SetTextColor(tcell.ColorWhite).
// 			SetAlign(tview.AlignCenter)
// 		buttonsTable.SetCell(0, i, btn)
// 	}

// 	updateButtonFocus := func(active int) {
// 		for i := 0; i < len(buttonLabels); i++ {
// 			cell := buttonsTable.GetCell(0, i)
// 			if i == active {
// 				cell.SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold)
// 			} else {
// 				cell.SetTextColor(tcell.ColorWhite).SetAttributes(tcell.AttrNone)
// 			}
// 		}
// 	}

// 	activeButton := 0
// 	buttonAction := func(i int) {
// 		tasks := tasksByFolder[currentFolder]
// 		switch i {
// 		case 0: // Run All
// 			for j := range tasks {
// 				tasks[j].Status = "Running"
// 			}
// 		case 1: // Stop All
// 			for j := range tasks {
// 				tasks[j].Status = "Ready"
// 			}
// 		case 2: // Enable All
// 			for j := range tasks {
// 				if tasks[j].Status == "Disabled" {
// 					tasks[j].Status = "Ready"
// 				}
// 			}
// 		case 3: // Disable All
// 			for j := range tasks {
// 				tasks[j].Status = "Disabled"
// 			}
// 		}
// 		tasksByFolder[currentFolder] = tasks
// 		refreshTable(currentFolder)
// 	}

// 	// ===== Flex layout =====
// 	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
// 		AddItem(
// 			tview.NewFlex().
// 				AddItem(tree, 30, 1, true).
// 				AddItem(table, 0, 3, false),
// 			0, 1, true).
// 		AddItem(buttonsTable, 3, 0, false)

// 	// ===== Enter на задаче =====
// 	table.SetSelectedFunc(func(row, column int) {
// 		if row == 0 {
// 			return
// 		}
// 		tasks := tasksByFolder[currentFolder]
// 		task := tasks[row-1]

// 		modal := tview.NewModal().
// 			SetText(fmt.Sprintf("Task: %s\nStatus: %s\nLast Result: %s", task.Name, task.Status, task.Result)).
// 			AddButtons([]string{"Run", "Stop", "Enable", "Disable", "Cancel"}).
// 			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
// 				switch buttonLabel {
// 				case "Run":
// 					tasks[row-1].Status = "Running"
// 				case "Stop":
// 					tasks[row-1].Status = "Ready"
// 				case "Enable":
// 					if tasks[row-1].Status == "Disabled" {
// 						tasks[row-1].Status = "Ready"
// 					}
// 				case "Disable":
// 					tasks[row-1].Status = "Disabled"
// 				}
// 				tasksByFolder[currentFolder] = tasks
// 				refreshTable(currentFolder)
// 				app.SetFocus(table)
// 				app.SetRoot(mainFlex, true)
// 			})
// 		app.SetRoot(modal, true)
// 	})

// 	// ===== Управление клавиатурой =====
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		focus := app.GetFocus()
// 		switch event.Key() {
// 		case tcell.KeyEsc:
// 			app.Stop()
// 		case tcell.KeyTab:
// 			if focus == tree {
// 				app.SetFocus(table)
// 			} else if focus == table {
// 				app.SetFocus(buttonsTable)
// 				activeButton = 0
// 				updateButtonFocus(activeButton)
// 			} else {
// 				app.SetFocus(tree)
// 			}
// 		case tcell.KeyRight:
// 			if focus == buttonsTable && activeButton < len(buttonLabels)-1 {
// 				activeButton++
// 				updateButtonFocus(activeButton)
// 			}
// 		case tcell.KeyLeft:
// 			if focus == buttonsTable && activeButton > 0 {
// 				activeButton--
// 				updateButtonFocus(activeButton)
// 			}
// 		case tcell.KeyEnter:
// 			if focus == buttonsTable {
// 				buttonAction(activeButton)
// 			}
// 		}
// 		return event
// 	})

// 	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
// 		panic(err)
// 	}
// }
