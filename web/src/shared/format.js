const locales = {
  BDT: "en-BD",
  INR: "en-IN",
  PKR: "en-PK",
  USD: "en-US",
};

export function money(value, currency = "BDT") {
  const amount = Number(value || 0);
  return `${amount.toLocaleString(locales[currency] || "en", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })} ${currency}`;
}

export function dateOnly(value) {
  if (!value) return "";
  return String(value).slice(0, 10);
}

export function monthLabel(year, month) {
  return new Date(year, month - 1, 1).toLocaleString("en", {
    month: "long",
    year: "numeric",
  });
}

export function accountTypeLabel(value) {
  switch (value) {
    case "cash":
      return "Cash";
    case "bank":
      return "Bank";
    case "mfs":
      return "MFS Wallet";
    default:
      return "Other";
  }
}
