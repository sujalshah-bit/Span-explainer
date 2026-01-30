import { Navigate } from "react-router-dom";
import { hasToken } from "../auth";
import type { JSX } from "react";

interface Props {
  children: JSX.Element;
}

export default function ProtectedRoute({ children }: Props) {
  if (!hasToken()) {
    return <Navigate to="/" replace />;
  }

  return children;
}