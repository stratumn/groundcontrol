// A tag that does nothing -- the tag is only added so that the code generator
// can find GraphQL queries.
export default function(literals: TemplateStringsArray, ...placeholders: string[]) {
  let result = "";

  for (let i = 0; i < placeholders.length; i++) {
    result += literals[i];
    result += placeholders[i];
  }

  result += literals[literals.length - 1];

  return result;
}
