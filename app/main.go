package main

import (

	// "time"
	"sync"

	"fyne.io/fyne/v2"

	// "fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/dialog"
	// "fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

type desktopUI struct {
	svc *Service
	app fyne.App
	win fyne.Window

	statusLabel      *widget.Label
	userLabel        *widget.Label
	friendHeader     *widget.Label
	messageHeader    *widget.Label
	messageEntry     *widget.Entry
	serverURLEntry   *widget.Entry
	newFriendEntry   *widget.Entry
	requestStatus    *widget.Label
	selectedFriend   *Friend
	selectedRequest  int
	friends          []Friend
	requests         []FriendRequest
	blocked          []BlockedUser
	messages         []ChatMessage
	friendList       *widget.List
	requestList      *widget.List
	blockedList      *widget.List
	messageList      *widget.List
	messageMu        sync.Mutex
	activeChatCancel func()
}

func main() {
	a := app.New()
	w := a.NewWindow("Hello World")

	w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()
}

func runDesktopApp() {

	// svc, err := NewService()
	// if err != nil {
	// panic(err)
	// }

	// guiApp := app.NewWithID("pochato.desktop")
	// window := guiApp.NewWindow("poCHATo")
	// window.Resize(fyne.NewSize(1280, 840))

	// fmt.Println("Hello")

	// ui := &desktopUI{
	// 	svc: svc,
	// 	app: guiApp,
	// 	win: window,
	// }

	// window.SetCloseIntercept(func() {
	// 	ui.svc.StopChat()
	// 	ui.app.Quit()
	// })

	// if authenticated, _ := svc.Bootstrap(); authenticated {
	// 	// ui.showMainScreen()
	// 	// ui.refreshAll()
	// } else {
	// 	// ui.showAuthScreen()
	// }

	// window.ShowAndRun()
}

// func (ui *desktopUI) showAuthScreen() {
// 	config := ui.svc.Config()
// 	ui.serverURLEntry = widget.NewEntry()
// 	ui.serverURLEntry.SetText(config.ServerURL)
// 	ui.serverURLEntry.SetPlaceHolder("http://localhost:8080")

// 	loginUsername := widget.NewEntry()
// 	loginUsername.SetPlaceHolder("Username")
// 	loginPassword := widget.NewPasswordEntry()
// 	loginPassword.SetPlaceHolder("Password")
// 	registerUsername := widget.NewEntry()
// 	registerUsername.SetPlaceHolder("Username")
// 	registerEmail := widget.NewEntry()
// 	registerEmail.SetPlaceHolder("Email")
// 	registerPassword := widget.NewPasswordEntry()
// 	registerPassword.SetPlaceHolder("Password")

// 	loginButton := widget.NewButtonWithIcon("Login", theme.LoginIcon(), func() {
// 		ui.runAsync(func() {
// 			if err := ui.applyServerURL(); err != nil {
// 				ui.showError(err)
// 				return
// 			}
// 			_, err := ui.svc.Login(strings.TrimSpace(loginUsername.Text), strings.TrimSpace(loginPassword.Text))
// 			if err != nil {
// 				ui.showError(err)
// 				return
// 			}
// 			ui.runOnMain(func() {
// 				ui.showMainScreen()
// 				ui.refreshAll()
// 			})
// 		})
// 	})

// 	registerButton := widget.NewButtonWithIcon("Register", theme.ContentAddIcon(), func() {
// 		ui.runAsync(func() {
// 			if err := ui.applyServerURL(); err != nil {
// 				ui.showError(err)
// 				return
// 			}
// 			_, err := ui.svc.Register(strings.TrimSpace(registerUsername.Text), strings.TrimSpace(registerEmail.Text), strings.TrimSpace(registerPassword.Text))
// 			if err != nil {
// 				ui.showError(err)
// 				return
// 			}
// 			ui.runOnMain(func() {
// 				ui.showMainScreen()
// 				ui.refreshAll()
// 			})
// 		})
// 	})

// 	form := container.NewVBox(
// 		widget.NewLabelWithStyle("poCHATo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
// 		widget.NewLabel("Secure desktop chat client"),
// 		widget.NewCard("Server", "Set the backend endpoint", ui.serverURLEntry),
// 		widget.NewCard("Login", "Use an existing account", container.NewVBox(loginUsername, loginPassword, loginButton)),
// 		widget.NewCard("Register", "Create a new account", container.NewVBox(registerUsername, registerEmail, registerPassword, registerButton)),
// 	)

// 	ui.win.SetContent(container.NewCenter(container.NewGridWrap(fyne.NewSize(460, 0), form)))
// }

// func (ui *desktopUI) showMainScreen() {
// 	ui.statusLabel = widget.NewLabel("Disconnected")
// 	ui.userLabel = widget.NewLabel("Signed out")
// 	ui.friendHeader = widget.NewLabel("Select a conversation")
// 	ui.messageHeader = widget.NewLabel("Messages")
// 	ui.requestStatus = widget.NewLabel("")

// 	ui.messageEntry = widget.NewEntry()
// 	ui.messageEntry.SetPlaceHolder("Type a message")
// 	ui.messageEntry.OnSubmitted = func(_ string) {
// 		ui.sendMessage(false)
// 	}

// 	ui.newFriendEntry = widget.NewEntry()
// 	ui.newFriendEntry.SetPlaceHolder("Friend username")

// 	ui.friendList = widget.NewList(
// 		func() int { return len(ui.friends) },
// 		func() fyne.CanvasObject { return widget.NewLabel("") },
// 		func(id widget.ListItemID, object fyne.CanvasObject) {
// 			label := object.(*widget.Label)
// 			if id >= 0 && id < len(ui.friends) {
// 				label.SetText(ui.shortID(ui.friends[id].FriendUserID))
// 			}
// 		},
// 	)
// 	ui.friendList.OnSelected = func(id widget.ListItemID) {
// 		if id >= 0 && id < len(ui.friends) {
// 			friend := ui.friends[id]
// 			ui.openFriend(friend)
// 		}
// 	}

// 	ui.requestList = widget.NewList(
// 		func() int { return len(ui.requests) },
// 		func() fyne.CanvasObject { return widget.NewLabel("") },
// 		func(id widget.ListItemID, object fyne.CanvasObject) {
// 			label := object.(*widget.Label)
// 			if id >= 0 && id < len(ui.requests) {
// 				label.SetText(fmt.Sprintf("From %s", ui.shortID(ui.requests[id].SenderID)))
// 			}
// 		},
// 	)
// 	ui.requestList.OnSelected = func(id widget.ListItemID) {
// 		ui.selectedRequest = int(id)
// 		if id >= 0 && id < len(ui.requests) {
// 			ui.requestStatus.SetText(fmt.Sprintf("Selected request from %s", ui.shortID(ui.requests[id].SenderID)))
// 		}
// 	}

// 	acceptRequest := widget.NewButtonWithIcon("Accept", theme.ContentAddIcon(), func() {
// 		ui.acceptSelectedRequest()
// 	})
// 	refreshRequests := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
// 		ui.refreshRequests()
// 	})

// 	ui.blockedList = widget.NewList(
// 		func() int { return len(ui.blocked) },
// 		func() fyne.CanvasObject { return widget.NewLabel("") },
// 		func(id widget.ListItemID, object fyne.CanvasObject) {
// 			label := object.(*widget.Label)
// 			if id >= 0 && id < len(ui.blocked) {
// 				label.SetText(ui.shortID(ui.blocked[id].BlockedUserID))
// 			}
// 		},
// 	)

// 	addFriendButton := widget.NewButtonWithIcon("Add Friend", theme.ContentAddIcon(), func() {
// 		username := strings.TrimSpace(ui.newFriendEntry.Text)
// 		if username == "" {
// 			return
// 		}
// 		ui.runAsync(func() {
// 			if err := ui.svc.AddFriend(username); err != nil {
// 				ui.showError(err)
// 				return
// 			}
// 			ui.runOnMain(func() {
// 				ui.newFriendEntry.SetText("")
// 				ui.refreshRequests()
// 			})
// 		})
// 	})

// 	friendPanel := container.NewBorder(
// 		container.NewVBox(
// 			widget.NewLabelWithStyle("Friends", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
// 			container.NewHBox(ui.newFriendEntry, addFriendButton),
// 		),
// 		nil,
// 		nil,
// 		nil,
// 		ui.friendList,
// 	)

// 	requestPanel := container.NewBorder(
// 		container.NewVBox(widget.NewLabelWithStyle("Requests", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), ui.requestStatus, container.NewHBox(acceptRequest, refreshRequests)),
// 		nil,
// 		nil,
// 		nil,
// 		ui.requestList,
// 	)

// 	blockedPanel := container.NewBorder(
// 		widget.NewLabelWithStyle("Blocked", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
// 		nil,
// 		nil,
// 		nil,
// 		ui.blockedList,
// 	)

// 	navTabs := container.NewAppTabs(
// 		container.NewTabItem("Friends", friendPanel),
// 		container.NewTabItem("Requests", requestPanel),
// 		container.NewTabItem("Blocked", blockedPanel),
// 	)

// 	ui.messageList = widget.NewList(
// 		func() int { return len(ui.messages) },
// 		func() fyne.CanvasObject {
// 			lbl1 := widget.NewLabel("")
// 			lbl2 := widget.NewLabel("")
// 			lbl2.Wrapping = fyne.TextWrapWord
// 			return container.NewVBox(lbl1, lbl2)
// 		},
// 		func(id widget.ListItemID, object fyne.CanvasObject) {
// 			if id < 0 || id >= len(ui.messages) {
// 				return
// 			}
// 			card := object.(*fyne.Container)
// 			title := card.Objects[0].(*widget.Label)
// 			body := card.Objects[1].(*widget.Label)
// 			message := ui.messages[id]
// 			if message.Incoming {
// 				title.SetText(fmt.Sprintf("%s · %s", ui.shortID(message.SenderID), message.CreatedAt.Format(time.Kitchen)))
// 			} else {
// 				title.SetText(fmt.Sprintf("You · %s", message.CreatedAt.Format(time.Kitchen)))
// 			}
// 			if message.IsHeart {
// 				body.SetText("❤️ " + message.Content)
// 			} else {
// 				body.SetText(message.Content)
// 			}
// 		},
// 	)

// 	sendButton := widget.NewButtonWithIcon("Send", theme.MailSendIcon(), func() {
// 		ui.sendMessage(false)
// 	})
// 	heartButton := widget.NewButtonWithIcon("Heart", theme.ConfirmIcon(), func() {
// 		ui.sendMessage(true)
// 	})
// 	messageControls := container.NewBorder(nil, nil, nil, container.NewHBox(heartButton, sendButton), ui.messageEntry)
// 	chatPane := container.NewBorder(
// 		container.NewVBox(ui.friendHeader, ui.statusLabel, ui.messageHeader),
// 		messageControls,
// 		nil,
// 		nil,
// 		ui.messageList,
// 	)

// 	split := container.NewHSplit(navTabs, chatPane)
// 	split.Offset = 0.30

// 	settingsButton := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
// 		ui.showSettingsDialog()
// 	})
// 	logoutButton := widget.NewButtonWithIcon("Logout", theme.LogoutIcon(), func() {
// 		ui.runAsync(func() {
// 			if err := ui.svc.Logout(); err != nil {
// 				ui.showError(err)
// 				return
// 			}
// 			ui.runOnMain(func() {
// 				ui.showAuthScreen()
// 			})
// 		})
// 	})
// 	ui.refreshUserBadge()
// 	toolbar := container.NewBorder(nil, nil, nil, container.NewHBox(settingsButton, logoutButton), ui.userLabel)
// 	ui.win.SetContent(container.NewBorder(toolbar, nil, nil, nil, split))
// }

// func (ui *desktopUI) refreshAll() {
// 	ui.refreshUserBadge()
// 	ui.refreshFriends()
// 	ui.refreshRequests()
// 	ui.refreshBlocked()
// }

// func (ui *desktopUI) refreshUserBadge() {
// 	if user, ok := ui.svc.CurrentUser(); ok {
// 		ui.userLabel.SetText(fmt.Sprintf("Signed in as %s", user.Username))
// 	} else {
// 		ui.userLabel.SetText("Signed out")
// 	}
// }

// func (ui *desktopUI) refreshFriends() {
// 	ui.runAsync(func() {
// 		friends, err := ui.svc.Friends()
// 		if err != nil {
// 			ui.showError(err)
// 			return
// 		}
// 		ui.runOnMain(func() {
// 			ui.friends = friends
// 			ui.friendList.Refresh()
// 		})
// 	})
// }

// func (ui *desktopUI) refreshRequests() {
// 	ui.runAsync(func() {
// 		requests, err := ui.svc.FriendRequests()
// 		if err != nil {
// 			ui.showError(err)
// 			return
// 		}
// 		ui.runOnMain(func() {
// 			ui.requests = requests
// 			ui.selectedRequest = -1
// 			ui.requestStatus.SetText("")
// 			ui.requestList.Refresh()
// 		})
// 	})
// }

// func (ui *desktopUI) refreshBlocked() {
// 	ui.runAsync(func() {
// 		blocked, err := ui.svc.BlockedUsers()
// 		if err != nil {
// 			ui.showError(err)
// 			return
// 		}
// 		ui.runOnMain(func() {
// 			ui.blocked = blocked
// 			ui.blockedList.Refresh()
// 		})
// 	})
// }

// func (ui *desktopUI) openFriend(friend Friend) {
// 	ui.selectedFriend = &friend
// 	ui.friendHeader.SetText(fmt.Sprintf("Conversation with %s", ui.shortID(friend.FriendUserID)))
// 	ui.statusLabel.SetText("Connecting...")
// 	ui.messages = nil
// 	ui.messageList.Refresh()

// 	ui.runAsync(func() {
// 		ui.runOnMain(func() {
// 			ui.messageHeader.SetText("Loading history...")
// 		})
// 		history, err := ui.svc.LoadHistory(friend, 50)
// 		if err == nil {
// 			ui.runOnMain(func() {
// 				ui.messages = history
// 				ui.messageHeader.SetText(fmt.Sprintf("Messages with %s", ui.shortID(friend.FriendUserID)))
// 				ui.messageList.Refresh()
// 			})
// 		}

// 		messages, status, err := ui.svc.StartChat(friend)
// 		if err != nil {
// 			ui.showError(err)
// 			return
// 		}

// 		go func() {
// 			for msg := range messages {
// 				msgCopy := msg

// 				ui.runOnMain(func() {
// 					ui.messages = append(ui.messages, msgCopy)
// 					ui.messageList.Refresh()
// 				})
// 			}
// 		}()

// 		go func() {
// 			for state := range status {
// 				text := state
// 				ui.runOnMain(func() {
// 					ui.statusLabel.SetText(text)
// 				})
// 			}
// 		}()
// 	})
// }

// func (ui *desktopUI) sendMessage(isHeart bool) {
// 	if ui.selectedFriend == nil {
// 		return
// 	}
// 	text := strings.TrimSpace(ui.messageEntry.Text)
// 	if text == "" && !isHeart {
// 		return
// 	}
// 	if isHeart && text == "" {
// 		text = "❤️"
// 	}

// 	friend := *ui.selectedFriend
// 	ui.runAsync(func() {
// 		if err := ui.svc.SendMessage(text, isHeart); err != nil {
// 			ui.showError(err)
// 			return
// 		}
// 		ui.runOnMain(func() {
// 			ui.messages = append(ui.messages, ChatMessage{
// 				SenderID:  "me",
// 				Content:   text,
// 				IsHeart:   isHeart,
// 				Incoming:  false,
// 				CreatedAt: time.Now(),
// 			})
// 			ui.messageEntry.SetText("")
// 			ui.friendHeader.SetText(fmt.Sprintf("Conversation with %s", ui.shortID(friend.FriendUserID)))
// 			ui.messageList.Refresh()
// 		})
// 	})
// }

// func (ui *desktopUI) acceptSelectedRequest() {
// 	if ui.selectedRequest < 0 || ui.selectedRequest >= len(ui.requests) {
// 		return
// 	}
// 	request := ui.requests[ui.selectedRequest]
// 	ui.runAsync(func() {
// 		if err := ui.svc.AcceptFriendRequest(request); err != nil {
// 			ui.showError(err)
// 			return
// 		}
// 		ui.runOnMain(func() {
// 			ui.refreshRequests()
// 			ui.refreshFriends()
// 			ui.refreshBlocked()
// 		})
// 	})
// }

// func (ui *desktopUI) showSettingsDialog() {
// 	config := ui.svc.Config()
// 	field := widget.NewEntry()
// 	field.SetText(config.ServerURL)
// 	info := widget.NewLabel(fmt.Sprintf("Data directory: %s", config.DataDir))
// 	content := container.NewVBox(field, info)
// 	dialog.ShowCustomConfirm("Settings", "Save", "Cancel", content, func(ok bool) {
// 		if !ok {
// 			return
// 		}
// 		ui.runAsync(func() {
// 			if err := ui.svc.UpdateServerURL(field.Text); err != nil {
// 				ui.showError(err)
// 				return
// 			}
// 			ui.runOnMain(func() {
// 				ui.statusLabel.SetText("Server URL saved. Re-login to use the new server.")
// 			})
// 		})
// 	}, ui.win)
// }

// func (ui *desktopUI) applyServerURL() error {
// 	if ui.serverURLEntry == nil {
// 		return nil
// 	}
// 	current := ui.svc.Config().ServerURL
// 	if strings.TrimSpace(ui.serverURLEntry.Text) == "" || strings.TrimSpace(ui.serverURLEntry.Text) == current {
// 		return nil
// 	}
// 	return ui.svc.UpdateServerURL(ui.serverURLEntry.Text)
// }

// func (ui *desktopUI) showError(err error) {
// 	if err == nil {
// 		return
// 	}
// 	ui.runOnMain(func() {
// 		dialog.ShowError(err, ui.win)
// 	})
// }

// func (ui *desktopUI) runAsync(fn func()) {
// 	go fn()
// }

// func (ui *desktopUI) runOnMain(fn func()) {
// 	fyne.Do(fn)
// }

// func (ui *desktopUI) shortID(value string) string {
// 	value = strings.TrimSpace(value)
// 	if len(value) <= 12 {
// 		return value
// 	}
// 	return value[:12] + "…"
// }
