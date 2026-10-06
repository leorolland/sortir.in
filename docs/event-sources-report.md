# Sources d'événements — rapport de veille (octobre 2026)

Objectif : identifier les sources à intégrer pour enrichir la base d'événements
de sortir.in sur toute la France.

**Ce qui est déjà collecté :** allevents.in, bobine, quefaire.paris (opendata
Paris), infomaniak.

**Taille du marché (contexte, données 2024 — CNM / ministère de la Culture) :**
~230 000 représentations de spectacle vivant/an, 65 M de spectateurs, 58 311
concerts déclarés, 1 356 festivals payants, 2,4 Md€ de billetterie.

## Vue d'ensemble

| Source | API ? | Accès | Events France | Effort | Priorité |
|---|---|---|---|---|---|
| **OpenAgenda** | REST v2 + agrégat national Opendatasoft | Clé gratuite (`oa_pk_`) ou 0 clé via ODS | **39 325 à venir** (mesuré le 06/10/2026), 1,19 M en base | Faible | ★★★ |
| **DATAtourisme** | REST v1 + endpoint `/entertainmentAndEvent` | Clé gratuite (inscription) | ~45 000 (12 % de ~375 k POI, déc. 2025) | Faible-moyen | ★★★ |
| **Ticketmaster Discovery** | REST v2 | Clé gratuite immédiate | 230 k+ mondial ; FR = concerts grands formats (arenas, ex-Fnac) | Faible | ★★ |
| **Shotgun** | Pas d'API publique de découverte ; API organisateur + scraping pages villes | — | Très actif électro/étudiant (Paris, Lyon, Bordeaux, Toulouse…) | Moyen (fragile) | ★★ |
| **Meetup** | GraphQL | OAuth, recherche d'événements **réservée aux réseaux Pro** | Paris surtout (groupes tech/expat) | Moyen-élevé | ★ |
| **Eventbrite** | v3 mais **Event Search API fermée depuis 2019** | Org/venue par ID uniquement | Inconnu (marché FR modeste) | — | ✗ |
| **Bandsintown / Songkick** | Artist-par-artist / **nouvelles clés en pause + licence payante** | Partenariat | Concerts seulement | — | ✗ |
| **Resident Advisor** | Pas d'API publique ; scrapers tiers payants (extractr, Apify) | — | Électro (Paris, Lyon…) | — | ✗ (option) |
| **Billetweb / Weezevent** | API côté organisateur (user+key), pas d'annuaire global | — | Billetterie FR | — | ✗ |
| **Sortiraparis, L'Officiel des spectacles** | Pas d'API (org data.gouv Sortiraparis : 0 API) | Scraping | L'Officiel : 3 000+ évts/semaine Paris-IDF (déjà couvert par quefaire) | Élevé | ✗ |
| **Apidae** | REST de diffusion, mais projet + clé via un membre territorial (pro) | Partenariat | SIT régionaux (ARA, Sud…) | Élevé | ✗ (plus tard) |

## Fiches détaillées

### 1. OpenAgenda — priorité n°1

La plateforme qu'utilisent déjà de nombreuses collectivités (Nantes, Bordeaux,
Rouen, Limoges, Est Ensemble…) pour leurs agendas publics. Licence ouverte.

Deux voies d'accès :

- **Agrégat national public** (0 clé) :
  `public.opendatasoft.com/explore/dataset/evenements-publics-openagenda` —
  1 279 120 événements dont **1 187 948 en France**, **39 325 à venir**
  au 06/10/2026. Top villes à venir : Nantes 1651, Bordeaux 1354, Paris 976,
  Toulouse 939, Rennes 922, Lille 717, Marseille 652. Géoloc, dates,
  catégories incluses.
- **API v2 directe** : <https://developers.openagenda.com> — clé `oa_pk_`
  gratuite, `GET /v2/agendas/{uid}/events` (size max 300), route transverse
  `/v2/events` en expérimentation (accès sur demande à support@openagenda.com).

Attention : les événements récurrents ont plusieurs `timings` → à éclater en
occurrences pour coller au modèle Event.

### 2. DATAtourisme — priorité n°2

La plateforme nationale des offices de tourisme (ADN Tourisme + État).
Ontologie sémantique, licence ouverte, couverture communale quasi totale.

- **API :** <https://api.datatourisme.fr/v1/docs> — clé gratuite après
  inscription ; endpoint dédié **`GET /entertainmentAndEvent`** (Fêtes et
  Manifestations), filtres `insee`, `geo_distance`, `geo_bounding`,
  `page_size` max 100–250. Exemple réel : 5 087 événements en Pays de la
  Loire.
- **Implémentation :** le collecteur interroge l'endpoint Elasticsearch
  anonyme de l'explorateur officiel (`POST
  https://explore.datatourisme.fr/api/_search`, celui du site
  explore.datatourisme.fr) : mêmes données, aucune clé, filtres `type =
  EntertainmentAndEvent` + `geo_distance` + `takesPlaceAt.endDate >=
  aujourd'hui`. L'API officielle reste l'alternative documentée si l'endpoint
  interne change.
- **Volume :** ~375 000 POI dont **12 % ≈ 45 000 événements** (concerts,
  expositions, marchés, fêtes…).
- **Attention :** modèle riche (mapping à prévoir) ; certains événements
  annuels/sans date précise → filtrer avant `Event.IsValid` (> 15 jours
  rejetés).

### 3. Ticketmaster Discovery API — priorité n°3

- **API :** <https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/>
  — clé gratuite instantanée, `countryCode=FR`, `latlong+radius`,
  `locale=fr-fr`, 230 k+ événements. Ticketmaster France (ex-Fnac
  Billetterie) couvre les gros concerts/spectacles, ~6 M d'acheteurs/an sur
  le réseau FR.
- **Intérêt :** complète la longue traîne d'OpenAgenda/DATAtourisme par le
  haut du marché.

### 4. Shotgun — forte valeur, plus risqué

Leader français de la billetterie électro/soirées étudiantes. API publique
seulement côté organisateur (`api.shotgun.live/tickets`,
`smartboard-api.shotgun.live/.../organizers/{id}/events`) ; découverte
publique via pages villes `shotgun.live/fr/cities/{ville}` (payload JSON dans
le front). Un scraper léger sur une liste de villes est faisable mais fragile
(pas de contrat d'API).

### 5. Meetup — possible mais bridé

GraphQL `api.meetup.com/gql-ext`, docs <https://www.meetup.com/graphql/> —
mais la recherche d'événements par zone est **gated derrière Meetup Pro** ;
sinon interrogation groupe par groupe (OAuth requis). 430 k événements/mois
dans le monde, forte présence Paris. À traiter seulement si l'API Pro refuse
et qu'on accepte le crawl par groupes.

## Plan d'action recommandé

1. **Collecteur OpenAgenda via l'agrégat ODS** (aucune clé, licence ouverte,
   ~39 k événements à venir sur tout le territoire) — le plus gros gain
   immédiat.
2. **Collecteur DATAtourisme** (aucune clé via l'explorateur Elasticsearch ;
   clé gratuite pour l'API officielle) — comble les villes moyennes et le
   tourisme.
3. **Collecteur Ticketmaster Discovery** (clé gratuite) — concerts grand
   format.
4. Ensuite selon retours : **Shotgun** (scraping villes) et **Meetup** (si
   OAuth/Pro débloqué).
5. Côté produit : prévoir le **mapping des kinds** (les sources ont des
   taxonomies riches → voir `pkg/application/kinds.go`) et la
   **déduplication inter-sources** (l'index unique `(name, begin, end)` ne
   prend pas les variantes de noms).

Point de vigilance transversal : ces sources couvrent toute la France alors
que `cmd/populate` travaille ville par ville — les collecteurs nationaux
s'intègrent mieux avec un mode bounding box ou un rayon par `CollectLocation`.
