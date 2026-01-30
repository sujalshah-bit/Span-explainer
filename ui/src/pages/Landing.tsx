import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { register } from "../api";

export default function Landing() {
  const navigate = useNavigate();

  useEffect(() => {
    if (!localStorage.getItem("token")) {
      register().catch(console.error);
    }
  }, []);

  return (
    <div className="h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white p-8 rounded-xl shadow-md text-center w-96">
        <h1 className="text-2xl font-bold mb-2">Span Explainer</h1>
        <p className="text-gray-600 mb-6">
          Upload OTLP traces and get AI-powered explanations for your spans.
        </p>
        <button
          onClick={() => navigate("/upload")}
          className="w-full bg-pink-600 text-white py-2 rounded-lg"
        >
          Get Started
        </button>
      </div>
    </div>
  );
}