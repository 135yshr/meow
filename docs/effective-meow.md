# Effective Meow

Idiomatic patterns and conventions for writing clean, consistent Meow code.

## Placeholder Names

Every language has its go-to stand-in names — the world at large reaches for
`foo`, `bar`, and `baz`. In Meow, we have our own cast of characters:

| Name | Role |
|------|------|
| **Nyantyu** | The first cat on the scene — your default placeholder |
| **Tyako** | The second — shows up when one name isn't enough |
| **Tyomusuke** | The third — for when the gang's all here |

Use them in examples, tests, and documentation whenever you need throwaway
names. Keeping these consistent makes `.nyan` code feel at home everywhere
in the project.

```nyan
kitty Cat {
  name: string
  age: int
}

nyan nyantyu = Cat("Nyantyu", 3)
nyan tyako = Cat("Tyako", 5)
nyan tyomusuke = Cat("Tyomusuke", 2)

nya(nyantyu)
nya(tyako)
nya(tyomusuke)
```
