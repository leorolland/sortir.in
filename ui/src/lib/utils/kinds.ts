export const KIND_LABELS: Record<string, string> = {
  concert: 'Concerts',
  theater: 'Théâtre',
  festival: 'Festivals',
  party: 'Soirées',
  karaoke: 'Karaoké',
  business: 'Professionnel',
  'food-drinks': 'Food & boissons',
  sports: 'Sports',
  exhibitions: 'Expositions',
  'health-wellness': 'Bien-être',
  circus: 'Cirque',
  workshop: 'Ateliers',
  'flea-market': 'Brocantes',
  solidarity: 'Solidarité'
};

export function kindLabel(kind: string): string {
  return KIND_LABELS[kind] ?? kind.charAt(0).toUpperCase() + kind.slice(1);
}
