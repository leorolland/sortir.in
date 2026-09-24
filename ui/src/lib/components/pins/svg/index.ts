import { createDefaultPinSVG } from './DefaultPinSVG';
import { createMoviePinSVG } from './MoviePinSVG';
import { createTheaterPinSVG } from './TheaterPinSVG';
import { createConcertPinSVG } from './ConcertPinSVG';
import { createFestivalPinSVG } from './FestivalPinSVG';
import { createPartyPinSVG } from './PartyPinSVG';
import { PIN_PIXEL_RATIO } from './PinShape';

// Modern flat palette (Apple system colors), all high-contrast with white glyphs
const PIN_COLORS = {
  default: '#3A3A3C',
  movie: '#FF9500',
  concert: '#FF2D55',
  festival: '#BF5AF2',
  theater: '#5E5CE6',
  party: '#0A84FF'
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
export { PIN_COLORS, PIN_PIXEL_RATIO };
