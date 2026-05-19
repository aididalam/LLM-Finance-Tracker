import { Bot, Paperclip, RefreshCw, Send, Trash2, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { api } from "../../shared/api.js";
import { money } from "../../shared/format.js";

const WELCOME = {
  role: "assistant",
  content: "Hi! Tell me about a transaction, ask for a report, or upload a receipt photo.",
};

export function ChatAssistant({ currency, open, onClose, onSaved }) {
  const [messages, setMessages] = useState([WELCOME]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [inlineOpts, setInlineOpts] = useState([]);
  const bottomRef = useRef(null);
  const fileRef = useRef(null);
  const textRef = useRef(null);

  // Load persisted history once on mount
  useEffect(() => {
    api("/api/v1/chat/history")
      .then((data) => {
        if (Array.isArray(data) && data.length > 0) {
          setMessages([WELCOME, ...data]);
        }
      })
      .catch(() => {});
  }, []);

  // Scroll to bottom on new messages
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages, loading]);

  // Focus input when chat opens
  useEffect(() => {
    if (open) {
      setTimeout(() => textRef.current?.focus(), 80);
    }
  }, [open]);

  // ── Send a message (user text or option value) ────────
  async function send(text) {
    const trimmed = (text || input).trim();
    if (!trimmed || loading) return;
    setInput("");
    setInlineOpts([]);

    const userMsg = { role: "user", content: trimmed };
    // history = everything except the static WELCOME bubble
    const history = messages.filter((m) => m !== WELCOME);
    push(userMsg);
    setLoading(true);

    try {
      const data = await api("/api/v1/chat", {
        method: "POST",
        body: JSON.stringify({ messages: [...history, userMsg] }),
      });

      push({ role: "assistant", content: data.reply || "…" });

      if (data.action === "confirm" && data.expense) {
        const wallets = data.wallet_options || [];
        if (wallets.length === 0) {
          // No wallets configured — save directly
          setInlineOpts([
            { label: "✅ Save", value: "_confirm_expense", type: "positive", _expense: data.expense },
            { label: "❌ Cancel", value: "_cancel_expense", type: "negative" },
          ]);
        } else if (wallets.length === 1 && wallets[0].account_type !== "bank") {
          // Single non-bank match — show save with pre-selected wallet
          setInlineOpts([
            {
              label: `✅ Save (${walletLabel(wallets[0])})`,
              value: "_save_with_wallet",
              type: "positive",
              _expense: data.expense,
              _wallet_id: wallets[0].id,
              _card_id: null,
            },
            { label: "❌ Cancel", value: "_cancel_expense", type: "negative" },
          ]);
        } else {
          // Multiple wallets or bank — show selection
          setInlineOpts([
            ...wallets.map((w) => ({
              label: walletLabel(w),
              value: "_select_wallet",
              type: "custom",
              _expense: data.expense,
              _wallet: w,
            })),
            { label: "❌ Cancel", value: "_cancel_expense", type: "negative" },
          ]);
        }
      } else if (data.action === "delete_select" && Array.isArray(data.expenses) && data.expenses.length) {
        setInlineOpts([
          ...data.expenses.map((exp) => ({
            label: `🗑 ${exp.description} — ${money(exp.amount, exp.currency)}`,
            value: "_delete_expense",
            type: "negative",
            _expense_id: exp.entry_id,
          })),
          { label: "Cancel", value: "_cancel_expense", type: "neutral" },
        ]);
      }
    } catch (err) {
      push({ role: "assistant", content: `⚠️ ${err.message}` });
    } finally {
      setLoading(false);
    }
  }

  // ── Receipt upload ────────────────────────────────────
  async function uploadReceipt(event) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;

    const isPDF = file.type === "application/pdf" || file.name.toLowerCase().endsWith(".pdf");
    push({ role: "user", content: isPDF ? `PDF receipt: ${file.name}` : `Receipt image: ${file.name}` });
    setLoading(true);
    setInlineOpts([]);

    try {
      const form = new FormData();
      form.append("receipt", file);
      const parsed = await api("/api/v1/receipt/parse", { method: "POST", body: form });

      if (!parsed?.is_expense) {
        push({ role: "assistant", content: parsed?.not_expense_reply || "Could not read that receipt." });
        return;
      }

      if (parsed.items?.length > 1) {
        const total = parsed.items.reduce((s, it) => s + Number(it.amount || 0), 0);
        push({
          role: "assistant",
          content: `Found ${parsed.items.length} items. Total: ${money(total, parsed.items[0]?.currency || currency)}`,
        });
        setInlineOpts([
          { label: "💾 Save all", value: "_save_all_receipt", type: "positive", _items: parsed.items },
          { label: "❌ Cancel", value: "_cancel_receipt", type: "negative" },
        ]);
        return;
      }

      const exp = parsed.items?.[0] || parsed;
      push({ role: "assistant", content: buildReceiptSummary(exp, currency) });
      setInlineOpts([
        { label: "✅ Save", value: "_save_receipt", type: "positive", _expense: exp },
        { label: "❌ Cancel", value: "_cancel_receipt", type: "negative" },
      ]);
    } catch (err) {
      push({ role: "assistant", content: `Receipt upload failed: ${err.message}` });
    } finally {
      setLoading(false);
    }
  }

  // ── Option chip clicked ───────────────────────────────
  async function handleOption(opt) {
    setInlineOpts([]);

    if (opt.value === "_cancel_receipt" || opt.value === "_cancel_expense") {
      push({ role: "assistant", content: "Cancelled." });
      return;
    }
    if (opt.value === "_save_receipt" && opt._expense) {
      await saveExpense(opt._expense);
      return;
    }
    if (opt.value === "_save_all_receipt" && opt._items) {
      await saveAllExpenses(opt._items);
      return;
    }
    if (opt.value === "_confirm_expense" && opt._expense) {
      await saveExpense(opt._expense);
      return;
    }
    if (opt.value === "_select_wallet" && opt._expense && opt._wallet) {
      if (opt._wallet.account_type === "bank") {
        await askDebitCard(opt._expense, opt._wallet.id);
      } else {
        await saveExpenseWithWallet(opt._expense, opt._wallet.id, null);
      }
      return;
    }
    if (opt.value === "_save_with_wallet" && opt._expense) {
      await saveExpenseWithWallet(opt._expense, opt._wallet_id, opt._card_id || null);
      return;
    }
    if (opt.value === "_delete_expense" && opt._expense_id) {
      await deleteExpenseByID(opt._expense_id);
      return;
    }

    // Fallback: send the value as a chat message (e.g. session options)
    await send(opt.value);
  }

  async function saveExpense(exp) {
    try {
      await api("/api/v1/chat/confirm", {
        method: "POST",
        body: JSON.stringify({
          expense: exp,
          wallet_id: exp.wallet_id || "",
          wallet_bank_debit_card_id: exp.wallet_bank_debit_card_id || null,
        }),
      });
      push({
        role: "assistant",
        content: `✅ Saved: ${money(exp.amount, exp.currency || currency)} – ${exp.description}`,
      });
      onSaved?.();
    } catch (err) {
      push({ role: "assistant", content: `Save failed: ${err.message}` });
    }
  }

  async function saveAllExpenses(items) {
    let saved = 0;
    for (const item of items) {
      await saveExpense(item);
      saved++;
    }
    push({ role: "assistant", content: `✅ Saved ${saved} receipt items.` });
  }

  async function saveExpenseWithWallet(exp, walletId, cardId) {
    await saveExpense({ ...exp, wallet_id: walletId || "", wallet_bank_debit_card_id: cardId || null });
  }

  async function askDebitCard(exp, walletId) {
    let cards = [];
    try {
      cards = await api(`/api/v1/accounts/${walletId}/debit-cards`);
    } catch (_) {}
    const opts = [
      {
        label: "🏦 No card",
        value: "_save_with_wallet",
        type: "positive",
        _expense: exp,
        _wallet_id: walletId,
        _card_id: null,
      },
    ];
    if (Array.isArray(cards)) {
      for (const card of cards) {
        opts.push({
          label: `💳 **** ${card.last_4_digit}`,
          value: "_save_with_wallet",
          type: "custom",
          _expense: exp,
          _wallet_id: walletId,
          _card_id: card.id,
        });
      }
    }
    opts.push({ label: "❌ Cancel", value: "_cancel_expense", type: "negative" });
    setInlineOpts(opts);
  }

  async function deleteExpenseByID(id) {
    try {
      await api(`/api/v1/expenses/${id}`, { method: "DELETE" });
      push({ role: "assistant", content: "🗑 Transaction deleted." });
      onSaved?.();
    } catch (err) {
      push({ role: "assistant", content: `Delete failed: ${err.message}` });
    }
  }

  // ── New chat session ──────────────────────────────────
  async function newChat() {
    await api("/api/v1/chat/history", { method: "DELETE" }).catch(() => {});
    setMessages([WELCOME]);
    setInlineOpts([]);
    setInput("");
  }

  async function clearHistory() {
    if (!window.confirm("Clear chat history?")) return;
    await newChat();
  }

  function push(msg) {
    setMessages((m) => [...m, msg]);
  }

  function handleKeyDown(e) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  }

  if (!open) return null;

  return (
    <>
      {/* Sidebar panel — visible on desktop as sibling, on mobile as overlay */}
      <div className={`chat-sidebar ${open ? "open" : ""}`}>
        {/* Header */}
        <header className="chat-header">
          <div className="chat-title-block">
            <Bot size={18} color="var(--accent-light)" />
            <div>
              <div className="chat-title">Finance Assistant</div>
              <div className="chat-subtitle">Draft · Confirm · Report</div>
            </div>
          </div>
          <div className="chat-header-actions">
            <button className="icon-btn" onClick={newChat} title="New chat">
              <RefreshCw size={14} />
            </button>
            <button className="icon-btn" onClick={clearHistory} title="Clear history">
              <Trash2 size={14} />
            </button>
            <button className="icon-btn" onClick={onClose} title="Close">
              <X size={16} />
            </button>
          </div>
        </header>

        {/* Messages */}
        <div className="chat-messages">
          {messages.map((msg, i) => (
            <ChatBubble key={i} msg={msg} />
          ))}

          {inlineOpts.length > 0 && !loading && (
            <div className="inline-options">
              {inlineOpts.map((opt, i) => (
                <button
                  key={i}
                  className={`inline-opt ${opt.type || "custom"}`}
                  onClick={() => handleOption(opt)}
                >
                  {opt.icon && <span>{opt.icon}</span>}
                  {opt.label}
                </button>
              ))}
            </div>
          )}

          {loading && (
            <div className="chat-loading">
              <div className="chat-dots">
                <span />
                <span />
                <span />
              </div>
              Thinking…
            </div>
          )}
          <div ref={bottomRef} />
        </div>

        {/* Composer */}
        <div className="chat-composer">
          <input
            ref={fileRef}
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp,application/pdf,.pdf"
            hidden
            onChange={uploadReceipt}
          />
          <div className="composer-row">
            <button
              className="composer-btn"
              type="button"
              onClick={() => fileRef.current?.click()}
              title="Upload receipt"
            >
              <Paperclip size={15} />
            </button>
            <textarea
              ref={textRef}
              className="composer-textarea"
              rows={1}
              value={input}
              placeholder="Describe a transaction…"
              onChange={(e) => {
                setInput(e.target.value);
                // Auto-grow
                e.target.style.height = "auto";
                e.target.style.height = `${Math.min(e.target.scrollHeight, 120)}px`;
              }}
              onKeyDown={handleKeyDown}
            />
            <button
              className="composer-btn send"
              type="button"
              onClick={() => send()}
              disabled={loading || !input.trim()}
              title="Send"
            >
              <Send size={15} />
            </button>
          </div>
        </div>
      </div>
    </>
  );
}

function ChatBubble({ msg }) {
  const isUser = msg.role === "user";
  return (
    <div className={`chat-bubble-row ${isUser ? "mine" : ""}`}>
      <div className={`chat-bubble ${isUser ? "user" : "bot"}`}>
        <p>{msg.content}</p>
      </div>
    </div>
  );
}

function buildReceiptSummary(exp, currency) {
  const amt = money(exp.amount, exp.currency || currency);
  const cat = exp.subcategory ? `${exp.category} / ${exp.subcategory}` : exp.category;
  return `Receipt: ${amt} – ${exp.description || "Transaction"} (${cat || "Uncategorized"}). Save?`;
}

function walletLabel(w) {
  const icons = { cash: "💵", bank: "🏦", mfs: "📱" };
  return `${icons[w.account_type] || "💳"} ${w.name}`;
}
