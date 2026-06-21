import { useState, type ReactNode } from 'react';
import { useNavigate } from 'react-router';

import { useAuth } from '../../context/AuthContext';

export function UserMenu({ actions }: { readonly actions?: ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);

  if (!user) return null;

  async function handleLogout() {
    setOpen(false);
    await logout();
    navigate('/login', { replace: true });
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        className="flex items-center gap-2 border border-pewter/20 px-3 py-2 text-xs font-semibold uppercase leading-none tracking-[0.15em] text-pewter transition-all duration-200 hover:border-pewter hover:text-champagne"
      >
        <span className="inline-flex h-6 w-6 items-center justify-center border border-pewter/40 bg-pewter/10 text-xs font-bold leading-none text-pewter">
          <span className="translate-x-px translate-y-[1px]">{user.name.charAt(0).toUpperCase()}</span>
        </span>
        <span className="hidden translate-y-[2px] sm:inline">{user.name}</span>
      </button>

      {open && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setOpen(false)}
          />
          <div className="absolute right-0 z-50 mt-2 w-64 border-2 border-pewter/20 bg-charcoal shadow-lg">
            <div className="border-b border-pewter/10 px-4 py-3">
              <p className="text-sm font-medium text-champagne">{user.name}</p>
              <p className="text-xs text-pewter">{user.email}</p>
            </div>
            {actions ? <div className="border-b border-pewter/10 p-2">{actions}</div> : null}
            <button
              type="button"
              onClick={(event) => {
                event.preventDefault();
                event.stopPropagation();
                void handleLogout();
              }}
              className="hud-text-button w-full px-4 text-left text-sm text-champagne/80 transition-colors hover:bg-ruby/10 hover:text-ruby"
            >
              Sign out
            </button>
          </div>
        </>
      )}
    </div>
  );
}
