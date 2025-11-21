import { createDefaultPinSVG } from './DefaultPinSVG';
import { createMoviePinSVG } from './MoviePinSVG';
import { createTheaterPinSVG } from './TheaterPinSVG';
import { createConcertPinSVG } from './ConcertPinSVG';
import { createFestivalPinSVG } from './FestivalPinSVG';
import { createPartyPinSVG } from './PartyPinSVG';

// Couleurs pour chaque type d'événement
const PIN_COLORS = {
  default: '#333333',
  movie: '#ff9800',
  concert: '#e91e63',
  festival: '#9c27b0',
  theater: '#673ab7',
  party: '#3f51b5'
};

// Génère tous les pins SVG avec leurs couleurs respectives
export const pinSVGs = {
  default: createDefaultPinSVG(PIN_COLORS.default),
  movie: createMoviePinSVG(PIN_COLORS.movie),
  concert: createConcertPinSVG(PIN_COLORS.concert),
  festival: createFestivalPinSVG(PIN_COLORS.festival),
  theater: createTheaterPinSVG(PIN_COLORS.theater),
  party: createPartyPinSVG(PIN_COLORS.party)
};

// Exporte les fonctions de création pour une utilisation personnalisée
export {
  createDefaultPinSVG,
  createMoviePinSVG,
  createTheaterPinSVG,
  createConcertPinSVG,
  createFestivalPinSVG,
  createPartyPinSVG
};

// Exporte les couleurs pour une utilisation externe
export { PIN_COLORS };
