package main

import (
	"errors"
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/db"
	"whatsapp-agent-tool/internal/db/todo"
	"whatsapp-agent-tool/internal/profiles"
)

// cmdTodo is the CLI for the ai database's task tables (migration 008). A todo's
// owner is the login the tool runs as (WA_PROFILE); --owner overrides it. Writes
// are gated on can_give_edit_instruction in requiredPermission.
func cmdTodo(args []string) {
	if len(args) == 0 {
		failErr("todo needs a subcommand: add|list|show|done|reopen|cancel|edit|move|assign|delete|remind|due|users")
	}
	sub, rest := args[0], args[1:]
	var err error
	switch sub {
	case "add":
		err = todoAdd(rest)
	case "list":
		err = todoList(rest)
	case "show":
		err = todoShow(rest)
	case "done":
		err = todoSetStatus(rest, todo.StatusDone)
	case "reopen":
		err = todoSetStatus(rest, todo.StatusOpen)
	case "cancel":
		err = todoSetStatus(rest, todo.StatusCancelled)
	case "edit":
		err = todoEdit(rest)
	case "move":
		err = todoMove(rest)
	case "assign":
		err = todoAssign(rest)
	case "delete":
		err = todoDelete(rest)
	case "remind":
		err = todoRemind(rest)
	case "due":
		err = todoDue(rest)
	case "users":
		err = todoUsers(rest)
	default:
		failErr("unknown todo subcommand " + strconv.Quote(sub))
	}
	if err != nil {
		db.ClosePool()
		failErr(err.Error())
	}
	db.ClosePool()
}

func todoAdd(args []string) error {
	title := flag(args, "title")
	if strings.TrimSpace(title) == "" {
		if pos := positionalArgs(args, valueFlags); len(pos) > 0 {
			title = pos[0]
		}
	}
	in := todo.CreateInput{
		Title:       title,
		Description: flag(args, "desc"),
		Priority:    flag(args, "priority"),
		DueAt:       flag(args, "due"),
		RequestedBy: flag(args, "requested-by"),
		ChatRef:     flag(args, "chat"),
	}
	owner := resolveOwner(flag(args, "owner"))
	in.OwnerProfile, in.OwnerDisplay, in.OwnerNumber = owner.profile, owner.display, owner.number
	if p, found := argInt(args, "parent"); found {
		in.ParentID = positivePtr(p)
	}
	id, err := todo.Create(in)
	if err != nil {
		return err
	}
	ok(pair("action", "add"), pair("id", id), pair("owner", owner.profile))
	return nil
}

func todoList(args []string) error {
	f := todo.ListFilters{
		Status:    flag(args, "status"),
		Priority:  flag(args, "priority"),
		Overdue:   has(args, "overdue"),
		DueBefore: flag(args, "due-before"),
		Limit:     intOr(argInt(args, "limit")),
	}
	if has(args, "all") {
		f.OwnerProfile = ""
	} else {
		f.OwnerProfile = resolveOwner(flag(args, "owner")).profile
	}
	if has(args, "roots") {
		f.OnlyRoots = true
	}
	if p, found := argInt(args, "parent"); found {
		f.ParentID = positivePtr(p)
	}
	if has(args, "tree") {
		return todoListTree(f)
	}
	todos, err := todo.List(f)
	if err != nil {
		return err
	}
	ok(pair("action", "list"), pair("count", len(todos)), pair("todos", todos))
	return nil
}

func todoShow(args []string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("show needs a todo id")
	}
	detail, err := todo.Get(id)
	if err != nil {
		return err
	}
	if detail == nil {
		return errStr("todo " + strconv.FormatInt(id, 10) + " not found")
	}
	ok(pair("action", "show"), pair("todo", detail.Todo),
		pair("children", detail.Children), pair("reminders", detail.Reminders), pair("history", detail.History))
	return nil
}

func todoSetStatus(args []string, status string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("this subcommand needs a todo id: whatsapp-tool todo <cmd> <id>")
	}
	if err := todo.SetStatus(id, status, resolveOwner("").profile); err != nil {
		return err
	}
	ok(pair("action", "status"), pair("id", id), pair("status", status))
	return nil
}

func todoEdit(args []string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("edit needs a todo id: whatsapp-tool todo edit <id> [--title ..]")
	}
	p := todo.UpdatePatch{}
	if has(args, "title") {
		v := flag(args, "title")
		p.Title = &v
	}
	if has(args, "desc") {
		v := flag(args, "desc")
		p.Description = &v
	}
	if has(args, "priority") {
		v := flag(args, "priority")
		p.Priority = &v
	}
	if has(args, "clear-due") {
		v := ""
		p.DueAt = &v
	} else if has(args, "due") {
		v := flag(args, "due")
		p.DueAt = &v
	}
	if has(args, "requested-by") {
		v := flag(args, "requested-by")
		p.RequestedBy = &v
	}
	if has(args, "chat") {
		v := flag(args, "chat")
		p.ChatRef = &v
	}
	changed, err := todo.Update(id, p)
	if err != nil {
		return err
	}
	ok(pair("action", "edit"), pair("id", id), pair("changed", changed))
	return nil
}

func todoMove(args []string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("move needs a todo id: whatsapp-tool todo move <id> --parent <id|none>")
	}
	raw := flag(args, "parent")
	if raw == "" {
		return errStr("move needs --parent <id|none>")
	}
	var parent *int64
	if raw != "none" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return errStr("--parent must be a todo id or 'none'")
		}
		parent = positivePtr(n)
	}
	if err := todo.Move(id, parent, resolveOwner("").profile); err != nil {
		return err
	}
	ok(pair("action", "move"), pair("id", id), pair("parent", parent))
	return nil
}

func todoAssign(args []string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("assign needs a todo id: whatsapp-tool todo assign <id> --owner <profile>")
	}
	profile := flag(args, "owner")
	if profile == "" {
		return errStr("assign needs --owner <profile>")
	}
	owner := resolveOwner(profile)
	ownerID, err := todo.Assign(id, owner.profile, owner.display, owner.number)
	if err != nil {
		return err
	}
	ok(pair("action", "assign"), pair("id", id), pair("ownerId", ownerID), pair("owner", owner.profile))
	return nil
}

func todoDelete(args []string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("delete needs a todo id: whatsapp-tool todo delete <id>")
	}
	affected, err := todo.Delete(id)
	if err != nil {
		return err
	}
	ok(pair("action", "delete"), pair("id", id), pair("deleted", affected))
	return nil
}

func todoRemind(args []string) error {
	if len(args) == 0 {
		return errStr("remind needs a subcommand: add|list|rm|sent")
	}
	switch args[0] {
	case "add":
		return todoRemindAdd(args[1:])
	case "list":
		return todoRemindList(args[1:])
	case "rm":
		return todoRemindStatus(args[1:], todo.ReminderCancelled)
	case "sent":
		return todoRemindStatus(args[1:], todo.ReminderSent)
	default:
		return errStr("unknown remind subcommand: add|list|rm|sent")
	}
}

func todoRemindAdd(args []string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("remind add needs a todo id: whatsapp-tool todo remind add <id> --at \"2026-10-20 09:30\"")
	}
	at := flag(args, "at")
	if at == "" {
		return errStr("remind add needs --at <date|datetime>")
	}
	rid, err := todo.AddReminder(id, at, flag(args, "channel"), flag(args, "note"))
	if err != nil {
		return err
	}
	ok(pair("action", "remind-add"), pair("id", rid), pair("todoId", id))
	return nil
}

func todoRemindList(args []string) error {
	f := todo.ReminderFilters{Status: flag(args, "status"), Limit: intOr(argInt(args, "limit"))}
	if id, found := argInt(args, "todo"); found {
		f.TodoID = positivePtr(id)
	}
	reminders, err := todo.ListReminders(f)
	if err != nil {
		return err
	}
	ok(pair("action", "remind-list"), pair("count", len(reminders)), pair("reminders", reminders))
	return nil
}

func todoRemindStatus(args []string, status string) error {
	id, found := firstPositionalInt(args)
	if !found {
		return errStr("remind " + status + " needs a reminder id")
	}
	if _, err := todo.SetReminderStatus(id, status); err != nil {
		return err
	}
	ok(pair("action", "remind-"+status), pair("id", id))
	return nil
}

func todoDue(args []string) error {
	due, err := todo.Due(flag(args, "before"))
	if err != nil {
		return err
	}
	ok(pair("action", "due"), pair("count", len(due)), pair("due", due))
	return nil
}

func todoUsers(args []string) error {
	users, err := todo.ListUsers()
	if err != nil {
		return err
	}
	ok(pair("action", "users"), pair("count", len(users)), pair("users", users))
	return nil
}

// ---------------------------------------------------------------- tree

// todoListTree returns the rows of `list --tree`, nested by parent_id.
func todoListTree(f todo.ListFilters) error {
	if f.Limit < 500 {
		f.Limit = 500
	}
	todos, err := todo.List(f)
	if err != nil {
		return err
	}
	ok(pair("action", "list"), pair("count", len(todos)), pair("tree", buildForest(todos)))
	return nil
}

type todoNode struct {
	Todo     todo.Todo   `json:"todo"`
	Children []*todoNode `json:"children"`
}

// buildForest nests the flat rows; a child whose parent is absent from the set
// (filtered out) is promoted to a root so nothing is lost.
func buildForest(todos []todo.Todo) []*todoNode {
	nodes := make(map[int64]*todoNode, len(todos))
	order := make([]int64, 0, len(todos))
	for _, t := range todos {
		nodes[t.ID] = &todoNode{Todo: t, Children: []*todoNode{}}
		order = append(order, t.ID)
	}
	roots := make([]*todoNode, 0)
	for _, id := range order {
		n := nodes[id]
		if n.Todo.ParentID != nil {
			if p, okParent := nodes[*n.Todo.ParentID]; okParent {
				p.Children = append(p.Children, n)
				continue
			}
		}
		roots = append(roots, n)
	}
	return roots
}

// ---------------------------------------------------------------- helpers

// ownerRef is the resolved login a todo belongs to.
type ownerRef struct {
	profile string
	display string
	number  string
}

// resolveOwner maps a profile slug (or "" = current login) to its identity, using
// profiles.json for the display name / number when the account is registered.
func resolveOwner(profile string) ownerRef {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		profile = config.WAProfile()
	}
	if e, found := profiles.Find(profile); found {
		display := e.Description
		if display == "" {
			display = e.Name
		}
		return ownerRef{profile: profile, display: display, number: e.Number}
	}
	return ownerRef{profile: profile}
}

// valueFlags are the --flag value pairs so positionalArgs skips their values.
var valueFlags = map[string]bool{
	"--title": true, "--desc": true, "--priority": true, "--due": true,
	"--requested-by": true, "--chat": true, "--owner": true, "--parent": true,
	"--status": true, "--limit": true, "--due-before": true, "--at": true,
	"--channel": true, "--note": true, "--todo": true,
}

func argInt(args []string, name string) (int64, bool) {
	v := flag(args, name)
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func intOr(n int64, found bool) int {
	if !found {
		return 0
	}
	return int(n)
}

func positivePtr(n int64) *int64 {
	if n <= 0 {
		return nil
	}
	return &n
}

func firstPositionalInt(args []string) (int64, bool) {
	for _, a := range positionalArgs(args, valueFlags) {
		if n, err := strconv.ParseInt(a, 10, 64); err == nil {
			return n, true
		}
	}
	return 0, false
}

func errStr(msg string) error { return errors.New(msg) }
