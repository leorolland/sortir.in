// Store for metadata
import { writable } from "svelte/store";

// Create a writable store for metadata
export const metadata = writable({
  title: "Sortir.in",
  headline: "Sortir.in - Trouvez les événements près de chez vous",
  description: "Sortir.in - Trouvez les événements près de chez vous",
});
