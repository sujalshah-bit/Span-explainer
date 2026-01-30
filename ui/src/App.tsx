import { BrowserRouter, Routes, Route } from "react-router-dom";
import Landing from "./pages/Landing";
import UploadTrace from "./pages/UploadTrace";
import ExplainSpan from "./pages/ExplainSpan";
import Navbar from "./components/Navbar.tsx";
import ProtectedRoute from "./components/Protected.tsx";

export default function App() {
  return (
    <BrowserRouter>
      <Navbar />
      <Routes>
        <Route path="/" element={<Landing />} />

        <Route
          path="/upload"
          element={
            <ProtectedRoute>
              <UploadTrace />
            </ProtectedRoute>
          }
        />

        <Route
          path="/explain"
          element={
            <ProtectedRoute>
              <ExplainSpan />
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}