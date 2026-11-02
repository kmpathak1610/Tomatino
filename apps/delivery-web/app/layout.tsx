import "./globals.css";
import Link from "next/link";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-gray-50 text-gray-900">
        <nav className="bg-white shadow-sm sticky top-0 z-10">
          <div className="max-w-4xl mx-auto px-4 py-3 flex items-center justify-between">
            <Link href="/" className="text-lg font-bold text-orange-600">FoodExpress Rider</Link>
            <div className="flex gap-4 text-sm">
              <Link href="/">Deliveries</Link>
              <Link href="/earnings">Earnings</Link>
              <Link href="/login">Login</Link>
            </div>
          </div>
        </nav>
        {children}
      </body>
    </html>
  );
}