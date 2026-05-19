import { useEffect, useState } from "react";
import { Alert, Button, Card, Col, Form, Modal, Row } from "react-bootstrap";
import { api } from "../../shared/api.js";

const ACCOUNT_TYPES = [
  { value: "cash", label: "Cash", icon: "💵" },
  { value: "bank", label: "Bank Account", icon: "🏦" },
  { value: "mfs", label: "MFS Wallet", icon: "📱" },
];

function typeLabel(value) {
  return ACCOUNT_TYPES.find((t) => t.value === value)?.label ?? value;
}

function typeIcon(value) {
  return ACCOUNT_TYPES.find((t) => t.value === value)?.icon ?? "🗂️";
}

const EMPTY_FORM = { name: "", account_type: "cash" };
const EMPTY_CARD_FORM = { last_4_digit: "", bank_name: "", account_number: "", branch: "" };

export function Wallet({ app }) {
  const [accounts, setAccounts] = useState([]);
  const [message, setMessage] = useState(null);
  const [showAdd, setShowAdd] = useState(false);
  const [editing, setEditing] = useState(null);
  const [deleting, setDeleting] = useState(null);
  const [form, setForm] = useState(EMPTY_FORM);
  const [saving, setSaving] = useState(false);

  // Debit card state
  const [cardWallet, setCardWallet] = useState(null); // wallet whose cards we're managing
  const [cards, setCards] = useState([]);
  const [cardForm, setCardForm] = useState(EMPTY_CARD_FORM);
  const [showCardForm, setShowCardForm] = useState(false);
  const [cardSaving, setCardSaving] = useState(false);
  const [cardMessage, setCardMessage] = useState(null);

  useEffect(() => {
    load();
  }, []);

  async function load() {
    try {
      const data = await api("/api/v1/accounts");
      setAccounts(Array.isArray(data) ? data : []);
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    }
  }

  function openAdd() {
    setForm(EMPTY_FORM);
    setShowAdd(true);
  }

  function openEdit(account) {
    setEditing(account);
  }

  async function openCards(account) {
    setCardWallet(account);
    setCards([]);
    setCardMessage(null);
    setShowCardForm(false);
    setCardForm(EMPTY_CARD_FORM);
    try {
      const data = await api(`/api/v1/accounts/${account.id}/debit-cards`);
      setCards(Array.isArray(data) ? data : []);
    } catch (err) {
      setCardMessage({ variant: "danger", text: err.message });
    }
  }

  async function submitAdd(e) {
    e.preventDefault();
    if (!form.name.trim()) {
      setMessage({ variant: "warning", text: "Account name is required." });
      return;
    }
    setSaving(true);
    try {
      await api("/api/v1/accounts", {
        method: "POST",
        body: JSON.stringify({
          name: form.name.trim(),
          account_type: form.account_type,
        }),
      });
      setShowAdd(false);
      setMessage({ variant: "success", text: "Account created." });
      load();
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    } finally {
      setSaving(false);
    }
  }

  async function submitEdit(e) {
    e.preventDefault();
    if (!editing?.name?.trim()) {
      setMessage({ variant: "warning", text: "Account name is required." });
      return;
    }
    setSaving(true);
    try {
      await api(`/api/v1/accounts/${editing.id}`, {
        method: "PUT",
        body: JSON.stringify({ name: editing.name.trim() }),
      });
      setEditing(null);
      setMessage({ variant: "success", text: "Account updated." });
      load();
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    } finally {
      setSaving(false);
    }
  }

  async function confirmDelete() {
    if (!deleting) return;
    try {
      await api(`/api/v1/accounts/${deleting.id}`, { method: "DELETE" });
      setDeleting(null);
      setMessage({ variant: "success", text: "Account removed." });
      load();
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    }
  }

  async function submitAddCard(e) {
    e.preventDefault();
    if (cardForm.last_4_digit.length !== 4 || !/^\d{4}$/.test(cardForm.last_4_digit)) {
      setCardMessage({ variant: "warning", text: "Last 4 digits must be exactly 4 numbers." });
      return;
    }
    if (cards.length === 0 && !cardForm.bank_name.trim()) {
      setCardMessage({ variant: "warning", text: "Bank name is required for the first card." });
      return;
    }
    if (cards.length === 0 && cardForm.account_number.trim().length < 4) {
      setCardMessage({ variant: "warning", text: "Account number must be at least 4 characters." });
      return;
    }
    setCardSaving(true);
    try {
      const payload = {
        last_4_digit: cardForm.last_4_digit,
        bank_name: cardForm.bank_name.trim() || "Bank",
        account_number: cardForm.account_number.trim() || "0000",
        branch: cardForm.branch.trim() || null,
      };
      const card = await api(`/api/v1/accounts/${cardWallet.id}/debit-cards`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      setCards((prev) => [...prev, card]);
      setCardForm(EMPTY_CARD_FORM);
      setShowCardForm(false);
      setCardMessage({ variant: "success", text: `Card •••• ${card.last_4_digit} added.` });
    } catch (err) {
      setCardMessage({ variant: "danger", text: err.message });
    } finally {
      setCardSaving(false);
    }
  }

  // group by type in the display order
  const grouped = ACCOUNT_TYPES.map((t) => ({
    ...t,
    items: accounts.filter((a) => a.account_type === t.value),
  })).filter((g) => g.items.length > 0);

  const needsBankDetails = cards.length === 0;

  return (
    <div className="page-stack">
      {message && (
        <Alert variant={message.variant} dismissible onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <div className="section-toolbar">
        <div>
          <div className="section-title">Accounts</div>
          <div className="muted-text">
            Manage your balance accounts — cash, bank, cards, wallets, and loans.
          </div>
        </div>
        <Button variant="primary" size="sm" onClick={openAdd}>
          + Add Account
        </Button>
      </div>

      {accounts.length === 0 && (
        <div className="empty-state">No accounts yet. Add your first account to start tracking balances.</div>
      )}

      {grouped.map((group) => (
        <div key={group.value}>
          <div className="section-label">
            {group.icon} {group.label}
          </div>
          <Row className="g-3">
            {group.items.map((account) => {
              const balance = account.balance ?? 0;
              const balanceColor = balance >= 0 ? "var(--bs-success)" : "var(--bs-danger)";
              return (
                <Col sm={6} lg={4} xl={3} key={account.id}>
                  <Card className="app-card h-100">
                    <Card.Body>
                      <div className="account-card-top">
                        <div className="account-type-icon">{typeIcon(account.account_type)}</div>
                      </div>
                      <div className="account-name">{account.name}</div>
                      <div className="account-balance" style={{ color: balanceColor }}>
                        {balance.toLocaleString(undefined, {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </div>
                      <div className="account-actions">
                        {account.account_type === "bank" && (
                          <Button variant="outline-secondary" size="sm" onClick={() => openCards(account)}>
                            💳 Cards
                          </Button>
                        )}
                        <Button variant="outline-secondary" size="sm" onClick={() => openEdit(account)}>
                          Edit
                        </Button>
                        <Button variant="outline-danger" size="sm" onClick={() => setDeleting(account)}>
                          Remove
                        </Button>
                      </div>
                    </Card.Body>
                  </Card>
                </Col>
              );
            })}
          </Row>
        </div>
      ))}

      {/* Add modal */}
      <Modal show={showAdd} onHide={() => setShowAdd(false)} centered>
        <Form onSubmit={submitAdd}>
          <Modal.Header closeButton>
            <Modal.Title>Add Account</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Name</Form.Label>
              <Form.Control
                value={form.name}
                placeholder="e.g. Main Bank, Bkash Wallet"
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Type</Form.Label>
              <Form.Select
                value={form.account_type}
                onChange={(e) => setForm((f) => ({ ...f, account_type: e.target.value }))}
              >
                {ACCOUNT_TYPES.map((t) => (
                  <option key={t.value} value={t.value}>
                    {t.icon} {t.label}
                  </option>
                ))}
              </Form.Select>
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="outline-secondary" onClick={() => setShowAdd(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={saving}>
              {saving ? "Saving…" : "Create"}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Edit modal */}
      <Modal show={Boolean(editing)} onHide={() => setEditing(null)} centered>
        <Form onSubmit={submitEdit}>
          <Modal.Header closeButton>
            <Modal.Title>Edit Account</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Name</Form.Label>
              <Form.Control
                value={editing?.name || ""}
                onChange={(e) => setEditing((a) => ({ ...a, name: e.target.value }))}
              />
            </Form.Group>
            <Form.Group>
              <Form.Label>Type</Form.Label>
              <Form.Control value={typeLabel(editing?.account_type)} readOnly disabled />
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="outline-secondary" onClick={() => setEditing(null)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={saving}>
              {saving ? "Saving…" : "Save"}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Delete confirm modal */}
      <Modal show={Boolean(deleting)} onHide={() => setDeleting(null)} centered size="sm">
        <Modal.Header closeButton>
          <Modal.Title>Remove Account</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          Remove <strong>{deleting?.name}</strong>? This will hide it from the list but won't delete any
          transaction history.
        </Modal.Body>
        <Modal.Footer>
          <Button variant="outline-secondary" onClick={() => setDeleting(null)}>
            Cancel
          </Button>
          <Button variant="danger" onClick={confirmDelete}>
            Remove
          </Button>
        </Modal.Footer>
      </Modal>

      {/* Debit cards modal */}
      <Modal show={Boolean(cardWallet)} onHide={() => setCardWallet(null)} centered>
        <Modal.Header closeButton>
          <Modal.Title>💳 Debit Cards — {cardWallet?.name}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {cardMessage && (
            <Alert
              variant={cardMessage.variant}
              dismissible
              className="mb-3"
              onClose={() => setCardMessage(null)}
            >
              {cardMessage.text}
            </Alert>
          )}

          {cards.length === 0 && !showCardForm && (
            <p className="text-muted small">No debit cards linked yet.</p>
          )}

          {cards.length > 0 && (
            <div className="mb-3">
              {cards.map((c) => (
                <div
                  key={c.id}
                  className="d-flex align-items-center gap-2 p-2 mb-2 rounded"
                  style={{ background: "var(--bs-secondary-bg)" }}
                >
                  <span>💳</span>
                  <span className="fw-semibold">•••• {c.last_4_digit}</span>
                  <small className="text-muted ms-auto">{c.is_active ? "Active" : "Inactive"}</small>
                </div>
              ))}
            </div>
          )}

          {showCardForm ? (
            <Form onSubmit={submitAddCard}>
              {needsBankDetails && (
                <>
                  <p className="text-muted small mb-2">
                    Enter your bank account details to link the first card.
                  </p>
                  <Form.Group className="mb-2">
                    <Form.Label>Bank Name</Form.Label>
                    <Form.Control
                      placeholder="e.g. Dutch-Bangla Bank"
                      value={cardForm.bank_name}
                      onChange={(e) => setCardForm((f) => ({ ...f, bank_name: e.target.value }))}
                    />
                  </Form.Group>
                  <Form.Group className="mb-2">
                    <Form.Label>Account Number</Form.Label>
                    <Form.Control
                      placeholder="e.g. 1234567890"
                      value={cardForm.account_number}
                      onChange={(e) => setCardForm((f) => ({ ...f, account_number: e.target.value }))}
                    />
                  </Form.Group>
                  <Form.Group className="mb-2">
                    <Form.Label>
                      Branch <span className="text-muted">(optional)</span>
                    </Form.Label>
                    <Form.Control
                      placeholder="e.g. Gulshan Branch"
                      value={cardForm.branch}
                      onChange={(e) => setCardForm((f) => ({ ...f, branch: e.target.value }))}
                    />
                  </Form.Group>
                </>
              )}
              <Form.Group className="mb-3">
                <Form.Label>Last 4 Digits of Card</Form.Label>
                <Form.Control
                  placeholder="e.g. 4321"
                  maxLength={4}
                  value={cardForm.last_4_digit}
                  onChange={(e) =>
                    setCardForm((f) => ({ ...f, last_4_digit: e.target.value.replace(/\D/g, "") }))
                  }
                />
              </Form.Group>
              <div className="d-flex gap-2">
                <Button type="submit" variant="primary" size="sm" disabled={cardSaving}>
                  {cardSaving ? "Saving…" : "Add Card"}
                </Button>
                <Button variant="outline-secondary" size="sm" onClick={() => setShowCardForm(false)}>
                  Cancel
                </Button>
              </div>
            </Form>
          ) : (
            <Button variant="outline-primary" size="sm" onClick={() => setShowCardForm(true)}>
              + Add Card
            </Button>
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="outline-secondary" onClick={() => setCardWallet(null)}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
}
