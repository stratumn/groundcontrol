import { ArraySliceMetaInfo, Connection, ConnectionArguments } from "graphql-relay";

// Modified version that uses IDs instead of offsets to paginate so that it works
// even if the array changes.
export function connectionFromArray<T>(
  data: T[],
  { after, before, first, last }: ConnectionArguments,
  getId: (obj: T) => string,
): Connection<T> {
  const beforeOffset = getOffsetWithDefault(data, getId, before, data.length);
  const afterOffset = getOffsetWithDefault(data, getId, after, -1);

  let startOffset = Math.max(afterOffset, -1) + 1;
  let endOffset = Math.min(data.length, beforeOffset);

  if (typeof first === "number") {
    if (first < 0) {
      throw new Error("Argument \"first\" must be a non-negative integer");
    }

    endOffset = Math.min(
      endOffset,
      startOffset + first,
    );
  }
  if (typeof last === "number") {
    if (last < 0) {
      throw new Error("Argument \"last\" must be a non-negative integer");
    }

    startOffset = Math.max(
      startOffset,
      endOffset - last,
    );
  }

  // If supplied slice is too large, trim it down before mapping over it.
  const slice = data.slice(
    Math.max(startOffset, 0),
    endOffset,
  );

  const edges = slice.map((value, index) => ({
    cursor: getId(data[startOffset + index]),
    node: value,
  }));

  const firstEdge = edges[0];
  const lastEdge = edges[edges.length - 1];
  const lowerBound = after ? (afterOffset + 1) : 0;
  const upperBound = before ? beforeOffset : data.length;
  return {
    edges,
    pageInfo: {
      endCursor: lastEdge ? lastEdge.cursor : null,
      hasNextPage:
        typeof first === "number" ? endOffset < upperBound : false,
      hasPreviousPage:
        typeof last === "number" ? startOffset > lowerBound : false,
      startCursor: firstEdge ? firstEdge.cursor : null,
    },
  };
}

function getOffsetWithDefault<T>(
  data: T[],
  getId: (obj: T) => string,
  cursor: string | null | undefined,
  defaultOffset: number,
 ): number {
   if (!cursor) {
     return defaultOffset;
   }

   for (let i = 0; i < data.length; i++) {
     if (getId(data[i]) === cursor) {
       return i;
     }
   }

   return defaultOffset;
}