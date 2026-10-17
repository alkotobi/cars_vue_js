package main

// voice-reply — the voice/call rule of client-reply.md §6 as PURE CODE: when
// the client's last message is a voice note, a call or a missed call, the only
// correct reply is one fixed sentence plus the AI disclosure line (AGENTS.md
// §12). No model is involved, so the wording can never drift and no question
// queue entries are created.
//
//	whatsapp-tool voice-reply "+213..."     # open the chat, verify, send the fixed line

import (
	"fmt"
	"strings"

	"whatsapp-agent-tool/internal/channel/whatsapp/wa"
)

// voiceReplyText is the §6 sentence, verbatim — never reworded.
const voiceReplyText = "لا أستطيع فهم الرسائل الصوتية بالعربية، من فضلك اكتب سؤالك نصيًا حتى أتمكن من مساعدتك."

func cmdVoiceReply(cli wa.Client, args []string) {
	target := ""
	for _, a := range args {
		if !strings.HasPrefix(a, "--") {
			target = a
			break
		}
	}
	if target == "" {
		failErr(`usage: whatsapp-tool voice-reply "<chat name or phone>"`)
	}
	if err := cli.Init(); err != nil {
		failErr("whatsapp not ready: " + err.Error())
	}
	defer cli.Close()

	// Verify the trigger before answering it — a stray command must not post
	// the voice-note line into an ordinary text conversation.
	okOpen, err := cli.OpenChat(target)
	if err != nil {
		failErr(err.Error())
	}
	if !okOpen {
		failErr("chat not found (or not a contact of this account): " + target)
	}
	msgs, err := cli.GetLastMessages(1)
	if err != nil {
		failErr(err.Error())
	}
	if len(msgs) == 0 {
		failErr("no messages read from: " + target)
	}
	last := msgs[len(msgs)-1]
	switch last.Type {
	case wa.MessageVoice, wa.MessageAudio, wa.MessageCall, wa.MessageMissed:
	default:
		failErr(fmt.Sprintf("last message is %q, not a voice/call — nothing sent", last.Type))
	}

	text := voiceReplyText + "\n" + wa.AiDisclosureLine
	sent, err := cli.SendMessage(text)
	if err != nil {
		failErr(err.Error())
	}
	if !sent {
		failErr("send failed (browser did not confirm) for: " + target)
	}
	ok(pair("action", "voice-reply"), pair("chat", target), pair("triggerType", string(last.Type)),
		pair("chars", len(text)))
}
