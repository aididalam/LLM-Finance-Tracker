import { Download, Search, Trash2, TrendingDown, TrendingUp } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  Alert,
  Badge,
  Button,
  ButtonGroup,
  Card,
  Col,
  Form,
  Modal,
  Row,
  Spinner,
  Table,
} from "react-bootstrap";

import { api, buildQuery } from "../../shared/api.js";
import { dateOnly, money } from "../../shared/format.js";

const PAGE_SIZE = 20;

function accountLabel(account) {
  const bal = (account.balance ?? 0).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return `${account.name} — ${bal}`;
}

const emptyIncome = {
  description: "",
  amount: "",
  fees: "",
  wallet_id: "",
};

const emptyExpense = {
  description: "",
  amount: "",
  fees: "",
  expense_date: new Date().toISOString().slice(0, 10),
  wallet_id: "",
  category_id: "",
  subcategory_id: "",
  subcategory: "",
};

export function Transactions({ app }) {
  const [rows, setRows] = useState([]);
  const [categories, setCategories] = useState([]);
  const [subcategories, setSubcategories] = useState([]);
  const [accounts, setAccounts] = useState([]);
  const [filters, setFilters] = useState({ q: "", from: "", to: "", offset: 0 });
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState(null);

  // which modal is open: null | 'income' | 'expense' | 'edit'
  const [modal, setModal] = useState(null);
  const [editingId, setEditingId] = useState("");
  const [income, setIncome] = useState(emptyIncome);
  const [expense, setExpense] = useState(emptyExpense);
  const [editEntry, setEditEntry] = useState({});

  const [saving, setSaving] = useState(false);
  const searchTimer = useRef(null);

  const currency = app?.settings?.currency || "USD";

  useEffect(() => {
    api("/api/v1/categories")
      .then((d) => setCategories(Array.isArray(d) ? d : []))
      .catch(() => {});
    loadAccounts();
  }, []);

  useEffect(() => {
    loadTransactions();
  }, [filters.offset, filters.from, filters.to]);

  useEffect(() => {
    clearTimeout(searchTimer.current);
    searchTimer.current = setTimeout(() => {
      setFilters((f) => ({ ...f, offset: 0 }));
      loadTransactions({ ...filters, offset: 0 });
    }, 250);
    return () => clearTimeout(searchTimer.current);
  }, [filters.q]);

  async function loadAccounts() {
    try {
      const data = await api("/api/v1/accounts");
      setAccounts(Array.isArray(data) ? data : []);
    } catch {}
  }

  async function loadTransactions(nextFilters = filters) {
    setLoading(true);
    try {
      const data = await api(
        `/api/v1/expenses${buildQuery({
          limit: PAGE_SIZE,
          offset: nextFilters.offset,
          q: nextFilters.q,
          from: nextFilters.from,
          to: nextFilters.to,
        })}`,
      );
      setRows(Array.isArray(data) ? data : []);
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    } finally {
      setLoading(false);
    }
  }

  async function loadSubcategories(categoryID, selectedID = "", target = "income") {
    if (!categoryID) {
      setSubcategories([]);
      return;
    }
    try {
      const data = await api(`/api/v1/categories/${categoryID}/subcategories`);
      setSubcategories(Array.isArray(data) ? data : []);
      if (selectedID) {
        const setter = target === "income" ? setIncome : target === "expense" ? setExpense : setEditEntry;
        setter((e) => ({ ...e, subcategory_id: selectedID }));
      }
    } catch {
      setSubcategories([]);
    }
  }

  function openIncome() {
    setIncome(emptyIncome);
    setModal("income");
  }

  function openExpense() {
    setExpense(emptyExpense);
    setSubcategories([]);
    setModal("expense");
  }

  function openEdit(row) {
    setEditEntry({
      description: row.description || "",
      amount: row.amount || "",
      fees: row.fees || "",
      expense_date: dateOnly(row.expense_datetime),
      category_id: row.category_id || "",
      subcategory_id: row.subcategory_id || "",
      subcategory: "",
    });
    setEditingId(row.id);
    setSubcategories([]);
    setModal("edit");
    if (row.category_id) loadSubcategories(row.category_id, row.subcategory_id, "edit");
  }

  // ── income submit ──────────────────────────────────────────────────────────
  async function submitIncome(e) {
    e.preventDefault();
    const amount = Number(income.amount);
    if (!amount || amount <= 0) {
      setMessage({ variant: "warning", text: "Enter a valid amount." });
      return;
    }
    if (!income.wallet_id) {
      setMessage({ variant: "warning", text: "Select a wallet to deposit into." });
      return;
    }
    setSaving(true);
    try {
      await api(`/api/v1/accounts/${income.wallet_id}/income`, {
        method: "POST",
        body: JSON.stringify({
          amount,
          fees: Number(income.fees) || 0,
        }),
      });
      setModal(null);
      setMessage({ variant: "success", text: "Income recorded." });
      await loadAccounts();
      app?.refresh();
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    } finally {
      setSaving(false);
    }
  }

  // ── expense submit ────────────────────────────────────────────────────────
  async function submitExpense(e) {
    e.preventDefault();
    const amount = Number(expense.amount);
    if (!amount || amount <= 0) {
      setMessage({ variant: "warning", text: "Enter a valid amount." });
      return;
    }
    if (!expense.wallet_id) {
      setMessage({ variant: "warning", text: "Select a wallet to pay from." });
      return;
    }
    if (!expense.category_id) {
      setMessage({ variant: "warning", text: "Select a category." });
      return;
    }
    setSaving(true);
    try {
      await api("/api/v1/expenses", {
        method: "POST",
        body: JSON.stringify({
          description: expense.description,
          amount,
          fees: Number(expense.fees) || 0,
          expense_date: expense.expense_date,
          wallet_id: expense.wallet_id,
          category_id: expense.category_id || null,
          subcategory_id: expense.subcategory ? null : expense.subcategory_id || null,
          subcategory: expense.subcategory || null,
        }),
      });
      setModal(null);
      setMessage({ variant: "success", text: "Expense saved." });
      await loadTransactions();
      await loadAccounts();
      app?.refresh();
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    } finally {
      setSaving(false);
    }
  }

  // ── edit submit ────────────────────────────────────────────────────────────
  async function submitEdit(e) {
    e.preventDefault();
    const amount = Number(editEntry.amount);
    if (!amount || amount <= 0) {
      setMessage({ variant: "warning", text: "Enter a valid amount." });
      return;
    }
    setSaving(true);
    try {
      await api(`/api/v1/expenses/${editingId}`, {
        method: "PUT",
        body: JSON.stringify({
          ...editEntry,
          amount,
          category_id: editEntry.category_id || null,
          subcategory_id: editEntry.subcategory ? null : editEntry.subcategory_id || null,
          subcategory: editEntry.subcategory || null,
        }),
      });
      setModal(null);
      setMessage({ variant: "success", text: "Transaction updated." });
      await loadTransactions();
      await loadAccounts();
      app?.refresh();
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    } finally {
      setSaving(false);
    }
  }

  async function deleteEntry(id) {
    if (!window.confirm("Delete this transaction?")) return;
    try {
      await api(`/api/v1/expenses/${id}`, { method: "DELETE" });
      setMessage({ variant: "success", text: "Transaction deleted." });
      await loadTransactions();
      await loadAccounts();
      app?.refresh();
    } catch (err) {
      setMessage({ variant: "danger", text: err.message });
    }
  }

  function pickAccount(walletID, setter) {
    setter((e) => ({ ...e, wallet_id: walletID }));
  }

  function updateCat(categoryID, setter, target) {
    setter((e) => ({ ...e, category_id: categoryID, subcategory_id: "", subcategory: "" }));
    loadSubcategories(categoryID, "", target);
  }

  const exportHref = useMemo(
    () => `/api/v1/expenses/export${buildQuery({ from: filters.from, to: filters.to })}`,
    [filters.from, filters.to],
  );

  // ── category sections ──────────────────────────────────────────────────────
  function CategoryFields({ entry, setter, target }) {
    return (
      <>
        <Col md={6}>
          <Form.Label>Category</Form.Label>
          <Form.Select value={entry.category_id} onChange={(e) => updateCat(e.target.value, setter, target)}>
            <option value="">None</option>
            {categories.map((c) => (
              <option key={c.category_id} value={c.category_id}>
                {c.icon} {c.name}
              </option>
            ))}
          </Form.Select>
        </Col>
        <Col md={6}>
          <Form.Label>Subcategory</Form.Label>
          <Form.Select
            value={entry.subcategory_id}
            disabled={!entry.category_id || !!entry.subcategory}
            onChange={(e) => setter((x) => ({ ...x, subcategory_id: e.target.value }))}
          >
            <option value="">None</option>
            {subcategories.map((s) => (
              <option key={s.subcategory_id} value={s.subcategory_id}>
                {s.name}
              </option>
            ))}
          </Form.Select>
        </Col>
        <Col md={12}>
          <Form.Label>
            New subcategory <span className="muted-text">(optional)</span>
          </Form.Label>
          <Form.Control
            value={entry.subcategory}
            placeholder="Type to create"
            onChange={(e) =>
              setter((x) => ({
                ...x,
                subcategory: e.target.value,
                subcategory_id: e.target.value ? "" : x.subcategory_id,
              }))
            }
          />
        </Col>
      </>
    );
  }

  return (
    <div className="page-stack">
      {message && (
        <Alert variant={message.variant} dismissible onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <Card className="app-card">
        <Card.Header>
          <div>
            <Card.Title>Transactions</Card.Title>
            <Card.Text>Manual entries, LLM drafts, and account movements.</Card.Text>
          </div>
          <div className="toolbar-actions">
            <Button variant="success" size="sm" onClick={openIncome}>
              <TrendingUp size={15} /> Income
            </Button>
            <Button variant="danger" size="sm" onClick={openExpense}>
              <TrendingDown size={15} /> Expense
            </Button>
            <Button as="a" href={exportHref} variant="outline-secondary" size="sm">
              <Download size={15} /> CSV
            </Button>
          </div>
        </Card.Header>

        <Card.Body>
          <Row className="g-2 align-items-center">
            <Col md={4}>
              <div className="search-control">
                <Search size={16} />
                <Form.Control
                  value={filters.q}
                  placeholder="Search transactions"
                  onChange={(e) => setFilters((f) => ({ ...f, q: e.target.value }))}
                />
              </div>
            </Col>
            <Col md={3}>
              <Form.Control
                type="date"
                value={filters.from}
                onChange={(e) => setFilters((f) => ({ ...f, from: e.target.value, offset: 0 }))}
              />
            </Col>
            <Col md={3}>
              <Form.Control
                type="date"
                value={filters.to}
                onChange={(e) => setFilters((f) => ({ ...f, to: e.target.value, offset: 0 }))}
              />
            </Col>
          </Row>
        </Card.Body>

        <div className="table-wrap">
          <Table responsive hover className="app-table mb-0">
            <thead>
              <tr>
                <th>Date</th>
                <th>Description</th>
                <th>Category</th>
                <th>Wallet</th>
                <th className="text-end">Amount</th>
                <th className="text-end">Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan="6" className="empty-cell">
                    <Spinner size="sm" /> Loading…
                  </td>
                </tr>
              ) : rows.length ? (
                rows.map((row) => (
                  <tr key={row.id}>
                    <td className="text-muted">{dateOnly(row.expense_datetime)}</td>
                    <td>
                      <div className="fw-semibold text-truncate table-main">{row.description}</div>
                      {row.subcategory_name && <div className="muted-text small">{row.subcategory_name}</div>}
                    </td>
                    <td className="text-muted">{row.category_name || "-"}</td>
                    <td className="text-muted">{row.wallet_name || "-"}</td>
                    <td className="text-end fw-bold text-danger">{money(row.amount, currency)}</td>
                    <td className="text-end">
                      <ButtonGroup size="sm">
                        <Button variant="outline-secondary" onClick={() => openEdit(row)}>
                          Edit
                        </Button>
                        <Button variant="outline-danger" onClick={() => deleteEntry(row.id)}>
                          <Trash2 size={14} />
                        </Button>
                      </ButtonGroup>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="6" className="empty-cell">
                    No transactions found.
                  </td>
                </tr>
              )}
            </tbody>
          </Table>
        </div>

        <Card.Footer className="d-flex justify-content-between align-items-center">
          <span className="muted-text">{rows.length} visible</span>
          <ButtonGroup>
            <Button
              variant="outline-secondary"
              disabled={filters.offset === 0}
              onClick={() => setFilters((f) => ({ ...f, offset: Math.max(0, f.offset - PAGE_SIZE) }))}
            >
              Previous
            </Button>
            <Button
              variant="outline-secondary"
              disabled={rows.length < PAGE_SIZE}
              onClick={() => setFilters((f) => ({ ...f, offset: f.offset + PAGE_SIZE }))}
            >
              Next
            </Button>
          </ButtonGroup>
        </Card.Footer>
      </Card>

      {/* ── Income modal ─────────────────────────────────────────────────── */}
      <Modal show={modal === "income"} onHide={() => setModal(null)} centered>
        <Form onSubmit={submitIncome}>
          <Modal.Header closeButton>
            <Modal.Title>
              <TrendingUp size={18} className="text-success me-2" />
              Record Income
            </Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Row className="g-3">
              <Col md={12}>
                <Form.Label>
                  Deposit into <span className="text-danger">*</span>
                </Form.Label>
                <Form.Select
                  value={income.wallet_id}
                  onChange={(e) => setIncome((x) => ({ ...x, wallet_id: e.target.value }))}
                  required
                >
                  <option value="">— select wallet —</option>
                  {accounts.map((a) => (
                    <option key={a.id} value={a.id}>
                      {accountLabel(a)}
                    </option>
                  ))}
                </Form.Select>
                {accounts.length === 0 && (
                  <Form.Text className="text-warning">
                    No wallets yet — add one in the Accounts tab first.
                  </Form.Text>
                )}
              </Col>
              <Col md={8}>
                <Form.Label>
                  Amount <span className="text-danger">*</span>
                </Form.Label>
                <Form.Control
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={income.amount}
                  onChange={(e) => setIncome((x) => ({ ...x, amount: e.target.value }))}
                  required
                />
              </Col>
              <Col md={4}>
                <Form.Label>Fees</Form.Label>
                <Form.Control
                  type="number"
                  step="0.01"
                  min="0"
                  value={income.fees}
                  onChange={(e) => setIncome((x) => ({ ...x, fees: e.target.value }))}
                />
              </Col>
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="outline-secondary" onClick={() => setModal(null)}>
              Cancel
            </Button>
            <Button type="submit" variant="success" disabled={saving}>
              {saving ? "Saving…" : "Record Income"}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* ── Expense modal ────────────────────────────────────────────────── */}
      <Modal show={modal === "expense"} onHide={() => setModal(null)} centered size="lg">
        <Form onSubmit={submitExpense}>
          <Modal.Header closeButton>
            <Modal.Title>
              <TrendingDown size={18} className="text-danger me-2" />
              Record Expense
            </Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Row className="g-3">
              <Col md={8}>
                <Form.Label>Description</Form.Label>
                <Form.Control
                  value={expense.description}
                  placeholder="e.g. Groceries, rent, electricity"
                  onChange={(e) => setExpense((x) => ({ ...x, description: e.target.value }))}
                />
              </Col>
              <Col md={4}>
                <Form.Label>Date</Form.Label>
                <Form.Control
                  type="date"
                  value={expense.expense_date}
                  onChange={(e) => setExpense((x) => ({ ...x, expense_date: e.target.value }))}
                />
              </Col>
              <Col md={6}>
                <Form.Label>
                  Amount <span className="text-danger">*</span>
                </Form.Label>
                <Form.Control
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={expense.amount}
                  onChange={(e) => setExpense((x) => ({ ...x, amount: e.target.value }))}
                  required
                />
              </Col>
              <Col md={6}>
                <Form.Label>Fees</Form.Label>
                <Form.Control
                  type="number"
                  step="0.01"
                  min="0"
                  value={expense.fees}
                  onChange={(e) => setExpense((x) => ({ ...x, fees: e.target.value }))}
                />
              </Col>
              <Col md={12}>
                <Form.Label>
                  Pay from <span className="text-danger">*</span>
                </Form.Label>
                <Form.Select
                  value={expense.wallet_id}
                  onChange={(e) => pickAccount(e.target.value, setExpense)}
                  required
                >
                  <option value="">— select wallet —</option>
                  {accounts.map((a) => (
                    <option key={a.id} value={a.id}>
                      {accountLabel(a)}
                    </option>
                  ))}
                </Form.Select>
              </Col>
              <CategoryFields entry={expense} setter={setExpense} target="expense" />
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="outline-secondary" onClick={() => setModal(null)}>
              Cancel
            </Button>
            <Button type="submit" variant="danger" disabled={saving}>
              {saving ? "Saving…" : "Save Expense"}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* ── Edit modal ───────────────────────────────────────────────────── */}
      <Modal show={modal === "edit"} onHide={() => setModal(null)} centered size="lg">
        <Form onSubmit={submitEdit}>
          <Modal.Header closeButton>
            <Modal.Title>Edit Expense</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Row className="g-3">
              <Col md={8}>
                <Form.Label>Description</Form.Label>
                <Form.Control
                  value={editEntry.description || ""}
                  onChange={(e) => setEditEntry((x) => ({ ...x, description: e.target.value }))}
                />
              </Col>
              <Col md={4}>
                <Form.Label>Date</Form.Label>
                <Form.Control
                  type="date"
                  value={editEntry.expense_date || ""}
                  onChange={(e) => setEditEntry((x) => ({ ...x, expense_date: e.target.value }))}
                />
              </Col>
              <Col md={6}>
                <Form.Label>
                  Amount <span className="text-danger">*</span>
                </Form.Label>
                <Form.Control
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={editEntry.amount || ""}
                  onChange={(e) => setEditEntry((x) => ({ ...x, amount: e.target.value }))}
                  required
                />
              </Col>
              <Col md={6}>
                <Form.Label>Fees</Form.Label>
                <Form.Control
                  type="number"
                  step="0.01"
                  min="0"
                  value={editEntry.fees || ""}
                  onChange={(e) => setEditEntry((x) => ({ ...x, fees: e.target.value }))}
                />
              </Col>
              <CategoryFields entry={editEntry} setter={setEditEntry} target="edit" />
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="outline-secondary" onClick={() => setModal(null)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={saving}>
              {saving ? "Saving…" : "Save"}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
}
