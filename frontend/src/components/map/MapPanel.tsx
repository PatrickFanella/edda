import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getMapData } from '../../api/map';
import type { MapLocationResponse, LocationConnectionResponse } from '../../api/types';
import { cn } from '../../lib/cn';

interface MapPanelProps {
  readonly campaignId: string;
  readonly className?: string;
}

interface RegionGroup {
  region: string;
  locations: MapLocationResponse[];
}

function groupByRegion(locations: MapLocationResponse[]): RegionGroup[] {
  const regionMap = new Map<string, MapLocationResponse[]>();

  for (const loc of locations) {
    const region = loc.region || 'Unknown Region';
    const existing = regionMap.get(region);
    if (existing) {
      existing.push(loc);
    } else {
      regionMap.set(region, [loc]);
    }
  }

  return Array.from(regionMap.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([region, locs]) => ({ region, locations: locs }));
}

function findConnectionsForLocation(
  locationId: string,
  connections: LocationConnectionResponse[],
  locations: MapLocationResponse[],
): { targetName: string; description: string; travelTime: string }[] {
  const locationNames = new Map(locations.map((l) => [l.id, l.name]));
  const results: { targetName: string; description: string; travelTime: string }[] = [];

  for (const conn of connections) {
    if (conn.from_location_id === locationId) {
      results.push({
        targetName: locationNames.get(conn.to_location_id) ?? 'Unknown',
        description: conn.description,
        travelTime: conn.travel_time,
      });
    } else if (conn.bidirectional && conn.to_location_id === locationId && conn.from_location_id) {
      results.push({
        targetName: locationNames.get(conn.from_location_id) ?? 'Unknown',
        description: conn.description,
        travelTime: conn.travel_time,
      });
    }
  }

  return results;
}

export function MapPanel({ campaignId, className }: MapPanelProps) {
  const { data, isPending, isError, error } = useQuery({
    queryKey: ['campaign', campaignId, 'map'],
    queryFn: () => getMapData(campaignId),
    enabled: campaignId.length > 0,
  });

  const regionGroups = useMemo(() => {
    if (!data?.locations) return [];
    return groupByRegion(data.locations);
  }, [data?.locations]);

  if (isPending) {
    return (
      <div className={cn('border border-jade/20 bg-charcoal p-6 text-sm text-champagne/70', className)}>
        Loading map data...
      </div>
    );
  }

  if (isError) {
    return (
      <div className={cn('border border-ruby/40 bg-ruby/10 p-6 text-sm text-ruby', className)}>
        {error instanceof Error ? error.message : 'Failed to load map data.'}
      </div>
    );
  }

  if (!data || data.locations.length === 0) {
    return (
      <div className={cn('flex min-h-48 flex-col items-center justify-center border border-dashed border-jade/15 bg-charcoal/50 px-6 text-center', className)}>
        <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-pewter/80">No locations discovered</p>
        <p className="mt-3 max-w-md text-sm leading-7 text-pewter">
          Explore the world to discover new locations. They will appear on the map as you travel.
        </p>
      </div>
    );
  }

  return (
    <div className={cn('space-y-6', className)}>
      {regionGroups.map((group) => (
        <section key={group.region} className="space-y-3">
          <h3 className="border-b border-jade/20 pb-2 font-heading text-sm font-semibold uppercase tracking-[0.2em] text-jade">
            {group.region}
          </h3>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {group.locations.map((location) => (
              <LocationCard
                key={location.id}
                location={location}
                connections={findConnectionsForLocation(location.id, data.connections, data.locations)}
              />
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}

function LocationCard({
  location,
  connections,
}: {
  readonly location: MapLocationResponse;
  readonly connections: { targetName: string; description: string; travelTime: string }[];
}) {
  return (
    <div className="border-2 border-jade/20 bg-midnight/20 p-4 transition-all duration-200 hover:border-jade/40">
      <div className="flex items-start justify-between gap-2">
        <h4 className="text-sm font-semibold text-champagne">{location.name}</h4>
        <div className="flex gap-1.5">
          {location.player_visited && (
            <span className="rounded-sm border border-jade/30 bg-jade/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-jade">
              Visited
            </span>
          )}
          {location.player_known && !location.player_visited && (
            <span className="rounded-sm border border-gold/30 bg-gold/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-gold">
              Known
            </span>
          )}
        </div>
      </div>

      <p className="mt-1 text-[11px] font-medium uppercase tracking-[0.15em] text-pewter">
        {location.location_type}
      </p>

      {location.description && (
        <p className="mt-2 text-xs leading-5 text-champagne/70">{location.description}</p>
      )}

      {connections.length > 0 && (
        <div className="mt-3 border-t border-white/5 pt-2">
          <p className="text-[10px] font-semibold uppercase tracking-[0.15em] text-pewter/60">Connections</p>
          <ul className="mt-1 space-y-1">
            {connections.map((conn) => (
              <li key={conn.targetName} className="flex items-center gap-1.5 text-[11px] text-champagne/60">
                <span className="text-jade/60">&rarr;</span>
                <span>{conn.targetName}</span>
                {conn.travelTime && (
                  <span className="text-pewter/50">({conn.travelTime})</span>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
