"use client";
import { useState } from "react";
import { api } from "@/lib/api";
import { useRouter } from "next/navigation";

export default function LoginPage() {
  const router = useRouter();
  const [phone, setPhone] = useState("");
  const [otp, setOtp] = useState("");
  const [sent, setSent] = useState(false);

  const sendOtp = async () => {
    await api("/auth/otp/send", { method: "POST", body: JSON.stringify({ phone }) });
    setSent(true);
  };
  const verify = async () => {
    const res: any = await api("/auth/otp/verify", { method: "POST", body: JSON.stringify({ phone, otp }) });
    localStorage.setItem("token", res.access_token);
    router.push("/");
  };

  return (
    <main className="max-w-md mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Login</h1>
      {!sent ? (
        <div className="space-y-3">
          <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="Phone number" className="w-full border rounded p-2" />
          <button onClick={sendOtp} className="w-full bg-orange-600 text-white py-2 rounded">Send OTP</button>
          <p className="text-xs text-gray-500">Dev OTP: 123456</p>
        </div>
      ) : (
        <div className="space-y-3">
          <input value={otp} onChange={(e) => setOtp(e.target.value)} placeholder="Enter OTP" className="w-full border rounded p-2" />
          <button onClick={verify} className="w-full bg-orange-600 text-white py-2 rounded">Verify</button>
        </div>
      )}
    </main>
  );
}