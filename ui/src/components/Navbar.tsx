import { Link, useLocation } from "react-router-dom";
import { hasToken } from "../auth";
import clsx from "clsx";

export default function Navbar() {
  const location = useLocation();
  const authenticated = hasToken();

  function navLink(
    to: string,
    label: string,
    disabled: boolean
  ) {
    return (
      <Link
        to={disabled ? "#" : to}
        className={clsx(
          "px-3 py-1 rounded-lg text-sm",
          location.pathname === to
            ? "bg-pink-100 text-pink-700"
            : "text-gray-700",
          disabled && "opacity-40 cursor-not-allowed"
        )}
        onClick={(e) => {
          if (disabled) e.preventDefault();
        }}
      >
        {label}
      </Link>
    );
  }

  return (
    <div className="border-b bg-white">
      <div className="max-w-5xl mx-auto px-4 py-3 flex justify-between items-center">
        <span className="font-semibold text-pink-600">
          Span Explainer
        </span>

        <div className="flex gap-2">
          {navLink("/", "Home", false)}
          {navLink("/upload", "Upload Trace", !authenticated)}
          {navLink("/explain", "Explain Span", !authenticated)}
        </div>
      </div>
    </div>
  );
}