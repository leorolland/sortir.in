import { writable } from 'svelte/store';

const STORAGE_KEY = 'kind-filter';

function loadInitial(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

function createKindFilterStore() {
  const { subscribe, set } = writable<string | null>(loadInitial());

  return {
    subscribe,
    set: (kind: string | null) => {
      try {
        if (kind === null) localStorage.removeItem(STORAGE_KEY);
        else localStorage.setItem(STORAGE_KEY, kind);
      } catch {
        // storage unavailable (private mode) — selection just won't persist
      }
      set(kind);
    }
  };
}

// Selected event kind shared by the Suggestions pane (chips) and the map
// (pin visibility). null means no filter.
export const kindFilter = createKindFilterStore();
