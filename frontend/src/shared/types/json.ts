export type JsonPrimitive = string | number | boolean | null;

export type JsonValue = JsonPrimitive | JsonObject | JsonArray;

export interface JsonObject {
  [key: string]: JsonValue;
}

export type JsonArray = JsonValue[];

export const isJsonObject = (value: JsonValue): value is JsonObject =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

export const isJsonArray = (value: JsonValue): value is JsonArray =>
  Array.isArray(value);
