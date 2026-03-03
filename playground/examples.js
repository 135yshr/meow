"use strict";

const MEOW_EXAMPLES = [
    {
        name: "Hello World",
        code: `nya("Hello, World!")`
    },
    {
        name: "Fibonacci",
        code: `meow fib(n int) int {
    sniff (n <= 1) {
        bring n
    }
    bring fib(n - 1) + fib(n - 2)
}

purr i (0..10) {
    nya(fib(i))
}`
    },
    {
        name: "FizzBuzz",
        code: `purr i (1..20) {
    sniff (i % 15 == 0) {
        nya("FizzBuzz")
    } scratch {
        sniff (i % 3 == 0) {
            nya("Fizz")
        } scratch {
            sniff (i % 5 == 0) {
                nya("Buzz")
            } scratch {
                nya(i)
            }
        }
    }
}`
    },
    {
        name: "List Operations",
        code: `nyan xs = [1, 2, 3, 4, 5]
nya("Original:", xs)

nyan doubled = lick(xs, paw(x int) { x * 2 })
nya("Doubled:", doubled)

nyan evens = picky(xs, paw(x int) { x % 2 == 0 })
nya("Evens:", evens)

nyan sum = curl(xs, 0, paw(acc int, x int) { acc + x })
nya("Sum:", sum)

# Pipe operator
nyan result = xs |=| picky(paw(x int) { x > 2 }) |=| lick(paw(x int) { x * 10 })
nya("Piped:", result)`
    },
    {
        name: "Kitty & Groom",
        code: `kitty Nyantyu {
    name: string
    age: int
}

groom Nyantyu {
    meow greet() string {
        bring "Meow! I am " + self.name
    }

    meow is_kitten() bool {
        bring self.age < 2
    }
}

nyan tama = Nyantyu("Tama", 3)
nyan chibi = Nyantyu("Chibi", 1)

nya(tama.greet())
nya(chibi.greet())
nya("Tama is kitten:", tama.is_kitten())
nya("Chibi is kitten:", chibi.is_kitten())`
    },
    {
        name: "Pattern Matching",
        code: `meow classify(n int) string {
    bring peek(n) {
        0 => "zero"
        1..9 => "single digit"
        10..99 => "double digit"
        _ => "big number"
    }
}

purr i (0..5) {
    nyan n = i * 25
    nya(n, "is", classify(n))
}`
    },
    {
        name: "Error Handling",
        code: `# gag catches errors and returns Furball
nyan result = gag(paw() { hiss("something broke") })
nya("Result:", result)
nya("Is error?", is_furball(result))

# ~> operator provides fallback
nyan safe = hiss("oops") ~> 42
nya("Safe value:", safe)

# ~> with function fallback
nyan handled = hiss("fail") ~> paw(err string) { "recovered" }
nya("Handled:", handled)`
    },
    {
        name: "Collar (Newtype)",
        code: `collar UserId = int
collar Email = string

nyan id = UserId(42)
nyan email = Email("tama@meow.cat")

nya("User ID:", id.value)
nya("Email:", email.value)`
    }
];
