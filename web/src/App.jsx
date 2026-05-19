import {
  BarChart3,
  Bot,
  LayoutDashboard,
  CreditCard,
  Landmark,
  PiggyBank,
  Tags,
  Settings,
  Sun,
  Moon,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { Wallet } from "./features/accounts/Wallet.jsx";
import { Budgets } from "./features/budgets/Budgets.jsx";
import { Categories } from "./features/categories/Categories.jsx";
import { ChatAssistant } from "./features/chat/ChatAssistant.jsx";
import { Dashboard } from "./features/dashboard/Dashboard.jsx";
import { SettingsPanel } from "./features/settings/SettingsPanel.jsx";
import { Transactions } from "./features/transactions/Transactions.jsx";
import { api } from "./shared/api.js";

const NAV = [
  { id: "dashboard", label: "Dashboard", icon: LayoutDashboard },
  { id: "transactions", label: "Transactions", icon: CreditCard },
  { id: "accounts", label: "Accounts", icon: Landmark },
  { id: "budgets", label: "Budgets", icon: PiggyBank },
  { id: "categories", label: "Categories", icon: Tags },
  { id: "settings", label: "Settings", icon: Settings },
];

export function App() {
  const [tab, setTab] = useState("dashboard");
  const [theme, setTheme] = useState(() => localStorage.getItem("theme") || "dark");
  const [chatOpen, setChatOpen] = useState(false);
  const [settings, setSettings] = useState({
    currency: "BDT",
    budget_alert_threshold: "80",
    budget_alert_enabled: "true",
  });
  const [refreshKey, setRefreshKey] = useState(0);

  // Apply theme
  useEffect(() => {
    document.documentElement.dataset.bsTheme = theme;
    localStorage.setItem("theme", theme);
    // Adjust CSS vars for light mode
    if (theme === "light") {
      const r = document.documentElement.style;
      r.setProperty("--bg", "#f5f4ff");
      r.setProperty("--sidebar-bg", "#ffffff");
      r.setProperty("--card-bg", "#ffffff");
      r.setProperty("--card-soft", "#f0eeff");
      r.setProperty("--border", "#ddd8ff");
      r.setProperty("--border-soft", "#ece9ff");
      r.setProperty("--text", "#1a1035");
      r.setProperty("--muted", "#6b5fa0");
      r.setProperty("--subtle", "#b0a8d8");
    } else {
      const r = document.documentElement.style;
      r.removeProperty("--bg");
      r.removeProperty("--sidebar-bg");
      r.removeProperty("--card-bg");
      r.removeProperty("--card-soft");
      r.removeProperty("--border");
      r.removeProperty("--border-soft");
      r.removeProperty("--text");
      r.removeProperty("--muted");
      r.removeProperty("--subtle");
    }
  }, [theme]);

  useEffect(() => {
    api("/api/v1/settings")
      .then((data) => data && setSettings((s) => ({ ...s, ...data })))
      .catch(() => {});
  }, []);

  const appCtx = useMemo(
    () => ({
      settings,
      setSettings,
      refresh: () => setRefreshKey((k) => k + 1),
    }),
    [settings],
  );

  const pageTitle = NAV.find((n) => n.id === tab)?.label ?? "Dashboard";

  function handleSaved() {
    appCtx.refresh();
  }

  return (
    <div className="app-shell">
      {/* ── Left sidebar ────────────────────────────── */}
      <aside className="sidebar">
        <div className="brand-block">
          <div className="brand-mark">
            <BarChart3 size={18} />
          </div>
          <div>
            <div className="brand-title">Expense Tracker</div>
            <div className="brand-subtitle">LLM-powered finance</div>
          </div>
        </div>

        <nav className="sidebar-nav">
          {NAV.map(({ id, label, icon: Icon }) => (
            <button key={id} className={`nav-item ${tab === id ? "active" : ""}`} onClick={() => setTab(id)}>
              <Icon size={16} />
              <span>{label}</span>
            </button>
          ))}
        </nav>

        <div className="sidebar-footer">v2.0 · LLM Finance</div>
      </aside>

      {/* ── Main content ────────────────────────────── */}
      <div className="main-panel">
        {/* Topbar */}
        <header className="topbar">
          <div className="topbar-left">
            <span className="topbar-page-title">{pageTitle}</span>
          </div>
          <div className="topbar-actions">
            <button
              className={`icon-btn ${chatOpen ? "active" : ""}`}
              onClick={() => setChatOpen((o) => !o)}
              title="Finance assistant"
            >
              <Bot size={16} />
            </button>
            <button
              className="icon-btn"
              onClick={() => setTheme((t) => (t === "dark" ? "light" : "dark"))}
              title="Toggle theme"
            >
              {theme === "dark" ? <Sun size={16} /> : <Moon size={16} />}
            </button>
          </div>
        </header>

        {/* Page content */}
        <div className="content-scroll">
          <div className="content-grid">
            {tab === "dashboard" && <Dashboard key={refreshKey} app={appCtx} />}
            {tab === "transactions" && <Transactions key={refreshKey} app={appCtx} />}
            {tab === "accounts" && <Wallet key={refreshKey} app={appCtx} />}
            {tab === "budgets" && <Budgets key={refreshKey} app={appCtx} />}
            {tab === "categories" && <Categories key={refreshKey} />}
            {tab === "settings" && <SettingsPanel app={appCtx} />}
          </div>
        </div>
      </div>

      {/* ── Chat sidebar panel ───────────────────────── */}
      <ChatAssistant
        currency={settings.currency || "BDT"}
        open={chatOpen}
        onClose={() => setChatOpen(false)}
        onSaved={handleSaved}
      />

      {/* ── Mobile: chat FAB (only shown when sidebar closed) ── */}
      {!chatOpen && (
        <button className="chat-fab" onClick={() => setChatOpen(true)} title="Open finance assistant">
          <Bot size={22} />
        </button>
      )}

      {/* ── Mobile bottom nav ────────────────────────── */}
      <nav className="bottom-nav">
        <div className="bottom-nav-inner">
          {NAV.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              className={`bottom-nav-item ${tab === id ? "active" : ""}`}
              onClick={() => setTab(id)}
            >
              <Icon size={20} />
              <span>{label}</span>
            </button>
          ))}
        </div>
      </nav>
    </div>
  );
}
