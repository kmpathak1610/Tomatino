"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function AdminDashboard() {
  const [stats, setStats] = useState<any>(null);
  const [error, setError] = useState("");
  useEffect(() => { api("/admin/stats").then(setStats).catch(() => setError("Login required")); }, []);

  return (
    <main className="max-w-6xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Marketplace Dashboard</h1>
      {error && <p className="text-red-600 mb-3">{error} <a href="/login" className="underline">Login</a></p>}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card label="Total Orders" value={stats?.total_orders} />
        <Card label="Customers" value={stats?.total_customers} />
        <Card label="Restaurants" value={stats?.total_restaurants} />
        <Card label="Delivery Partners" value={stats?.total_delivery_partners} />
        <Card label="Gross Order Value" value={stats ? `₹${stats.gross_order_value}` : undefined} />
      </div>
    </main>
  );
}

function Card({ label, value }: { label: string; value: any }) {
  return <div className="bg-white rounded shadow p-4"><div className="text-sm text-gray-500">{label}</div><div className="text-2xl font-bold">{value ?? "-"}</div></div>;
}