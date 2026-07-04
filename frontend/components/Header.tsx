"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";

export type Status = "ready" | "busy" | "error";

interface HeaderProps {
  status?: Status;
  statusText?: string;
}

export default function Header({ status, statusText }: HeaderProps) {
  const pathname = usePathname();

  return (
    <header className="header">
      <div className="brand">
        <span className="brand__mark">
          <Image src="/logo.png" alt="" width={26} height={26} priority />
        </span>
        <span className="brand__name">
          Kasi<span className="brand__accent">Delivery</span>
        </span>
      </div>
      <nav className="header__nav" aria-label="Primary">
        <Link className={`nav-link ${pathname === "/" ? "nav-link--active" : ""}`} href="/">
          Dispatch
        </Link>
        <Link className={`nav-link ${pathname === "/test-route" ? "nav-link--active" : ""}`} href="/test-route">
          Test Route
        </Link>
        {status && (
          <span className={`status-pill status-pill--${status}`} role="status" aria-live="polite">
            <span className="status-pill__dot" />
            <span>{statusText}</span>
          </span>
        )}
      </nav>
    </header>
  );
}
