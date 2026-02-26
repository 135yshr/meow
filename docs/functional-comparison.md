# Haskell / Elm との機能比較

Meow の関数型プログラミング機能を Haskell / Elm と比較し、現在の対応状況と今後の拡張候補を整理する。

## 1. 型システム

| 機能 | Haskell/Elm | Meow の現状 |
|------|-------------|-------------|
| 型推論 (Hindley-Milner) | 完全な型推論。型注釈なしでもコンパイル可 | 部分的。段階的型付け (Gradual Typing)。型注釈付き関数はネイティブ Go コードを生成 |
| ジェネリクス（多相型） | `map :: (a -> b) -> [a] -> [b]` | なし。全て `meowrt.Value` でボクシング |
| 型クラス / トレイト | `class Eq a where (==) :: a -> a -> Bool` | `trick`（インターフェース定義）+ `learn`（メソッド実装）。構造的型付け。デフォルト実装・ジェネリクスは未対応 |
| 代数的データ型 (ADT) | `data Maybe a = Nothing \| Just a` | `kitty`（Product 型）のみ。Sum 型なし |
| 型エイリアス | `type Name = String` | `breed Name = string`（透過的エイリアス。前方参照・チェーン可） |
| Newtype | `newtype Age = Age Int` | `collar Age = int`（名前的型付け。コンストラクタ `Age(42)` と `.value` アクセサ） |

### breed / collar の詳細

**breed（型エイリアス）** は Haskell の `type` に対応する透過的エイリアス。基底型と完全に互換で、条件式 (`sniff`)、ループ (`purr`)、算術演算、単項演算のすべてで基底型として振る舞う。前方参照やチェーン（`breed A = B` / `breed B = int`）にも対応。

```meow
breed Score = int
breed Flag = bool
nyan s Score = 42
nyan f Flag = yarn
sniff (f) { nya(s + 1) }    # Score は int として演算可
```

**collar（Newtype）** は Haskell の `newtype` に対応する名前的型。同じ基底型でも名前が異なれば別の型として扱われる。内部値には `.value` でアクセスする。

```meow
collar UserId = int
collar Email = string
nyan id = UserId(42)
nyan email = Email("nyantyu@meow.cat")
nya(id.value)                # => 42
nyan a = UserId(1)
nyan b = UserId(1)
judge(a == b, "same collar, same value")
```

### trick / learn の詳細

**trick（インターフェース）** は Haskell の `class` / Go の `interface` に対応する。メソッドシグネチャの集合を定義する。型が trick を満たすかは構造的に判定される（明示的な宣言は不要）。

```meow
trick Showable {
    meow show() string
}
```

**learn（メソッド実装）** は Haskell の `instance` / Go のメソッドレシーバに対応する。`kitty` や `collar` 型にメソッドを追加する。メソッド本体では `self` で自身のインスタンスを参照する。

```meow
kitty Cat { name: string, age: int }

learn Cat {
    meow show() string {
        bring self.name + " (age " + to_string(self.age) + ")"
    }
    meow is_kitten() bool {
        bring self.age < 1
    }
}

nyan c = Cat("Nyantyu", 3)
nya(c.show())       # => Nyantyu (age 3)
nya(c.is_kitten())  # => hairball
```

`collar` 型にもメソッドを追加できる。内部値には `self.value` でアクセスする。

```meow
collar Label = string

learn Label {
    meow display() string {
        bring "[ " + self.value + " ]"
    }
}

nyan tag = Label("important")
nya(tag.display())   # => [ important ]
```

Haskell の型クラスとの主な違い：
- **構造的型付け**: Haskell は `instance Eq Cat where ...` と明示的に宣言するが、Meow は必要なメソッドを `learn` で定義すれば自動的に trick を満たす
- **デフォルト実装なし**: Haskell の `class` ではデフォルト実装を定義できるが、Meow の `trick` はシグネチャのみ
- **ジェネリクスなし**: `trick Functor f where fmap :: (a -> b) -> f a -> f b` のような型パラメータ付き trick は未対応

### 型付き関数とネイティブコード生成

全パラメータと戻り値にプリミティブ型注釈（`int`, `float`, `string`, `bool`）が付いた関数は、Go のネイティブ型（`int64`, `float64`, `string`, `bool`）で直接コード生成される。ボクシング／アンボクシングのオーバーヘッドがなく、Go と同等のパフォーマンスで動作する。

```meow
# ネイティブ int64 演算にコンパイルされる
meow add(a int, b int) int {
    bring a + b
}
```

## 2. パターンマッチング

| 機能 | Haskell/Elm | Meow の現状 |
|------|-------------|-------------|
| 構造の分解 | `case list of (x:xs) -> ...` | なし。リスト・構造体の分解不可 |
| ネストパターン | `Just (Left x) -> ...` | なし |
| ガード条件 | `f x \| x > 0 = ...` | なし |
| OR パターン | Elm: `Red \| Blue -> "color"` | なし |
| as パターン | `node@(Node l r) -> ...` | なし |

Meow の `peek` はリテラル、範囲 (`1..10`)、ワイルドカード (`_`) に対応しているが、構造の分解やガード条件は未対応。

## 3. 関数型プログラミングの中核機能

| 機能 | Haskell/Elm | Meow の現状 |
|------|-------------|-------------|
| カリー化 | デフォルトで全関数がカリー化 | なし（手動でラムダのネストは可能） |
| 部分適用 | `add 1` で 1 引数関数を生成 | なし |
| 関数合成 | `f . g` や `f >> g` | `\|=\|` パイプのみ（合成演算子なし） |
| ポイントフリースタイル | `sum = foldr (+) 0` | 不可 |
| リスト内包表記 | `[x*2 \| x <- [1..10], even x]` | なし（`lick` + `picky` で代替） |
| 遅延評価 | Haskell: デフォルトで遅延評価 | なし（全て正格評価） |

### Meow が対応済みの関数型機能

- `lick` (map) / `picky` (filter) / `curl` (reduce) — 高階関数
- `paw` — ラムダ式（型注釈付きパラメータが必須）
- `|=|` — パイプ演算子（左辺の値を右辺の関数の第1引数に渡す）
- `~>` — エラー回復演算子（左辺が例外を投げた場合に右辺にフォールバック）
- `gag` — 例外捕捉（Go の `recover` に対応）
- `hiss` — 例外送出（Go の `panic` に対応。`"Hiss! "` プレフィックス付き）

## 4. モジュール・名前空間

| 機能 | Haskell/Elm | Meow の現状 |
|------|-------------|-------------|
| ユーザー定義モジュール | `module Foo exposing (..)` | なし。`fetch` で組み込みのみ |
| 選択的インポート | `import List exposing (map, filter)` | なし。パッケージ全体をインポート |
| qualified import | `import qualified Map as M` | なし |

## 5. 不変性・純粋性

| 機能 | Haskell/Elm | Meow の現状 |
|------|-------------|-------------|
| 参照透過性の保証 | コンパイラが副作用を追跡 | なし。副作用の追跡なし |
| 純粋関数の強制 | Haskell: IO モナドで分離 / Elm: Cmd/Sub | なし |
| 不変変数の宣言 | 全変数がデフォルト不変 | 全変数がデフォルト不変（同一スコープでの再宣言はエラー。内側スコープでのシャドーイングは可） |

## 6. 高度な型システム機能

| 機能 | Haskell/Elm | Meow の現状 |
|------|-------------|-------------|
| モナド | `Maybe`, `Either`, `IO` など | なし（`~>` が部分的に類似） |
| ファンクタ / Applicative | `fmap`, `<*>` | なし |
| 高カインド型 | `Functor f => f a -> f b` | なし |
| レコード更新構文 | Elm: `{ model \| count = 1 }` | なし |
| Phantom Type | `data Tagged tag a = Tagged a` | なし |

## 7. その他

| 機能 | Haskell/Elm | Meow の現状 |
|------|-------------|-------------|
| where / let ... in | ローカルスコープの定義 | なし |
| do 記法 | モナドの糖衣構文 | なし |
| タプル | `(1, "hello", True)` | なし（`kitty` で名前付きフィールドのみ） |
| カスタム演算子定義 | Haskell: `(+++) a b = ...` | なし |
| テストフレームワーク | 外部ライブラリ (HUnit, QuickCheck) | 組み込み。`judge`/`expect`/`refuse` + `test_` / `catwalk_` 規約 |

## 現在の型システムの対応状況まとめ

```
        ┌───────────────────────────────────────────────────────┐
        │                  Meow 型システム                        │
        ├──────────────┬──────────────┬────────────┬────────────┤
        │ プリミティブ  │ ユーザー定義  │ 型修飾子    │ 型クラス    │
        │ int          │ kitty(構造体)│ breed(透過) │ trick(IF)  │
        │ float        │              │ collar(名前)│ learn(実装) │
        │ string       │              │            │ self(参照)  │
        │ bool         │              │            │            │
        │ list         │              │            │            │
        │ furball      │              │            │            │
        └──────────────┴──────────────┴────────────┴────────────┘
                │               │               │
                ▼               ▼               ▼
        ┌───────────────┐ ┌──────────────┐ ┌──────────────────┐
        │ 型注釈あり     │ │ 前方参照      │ │ 構造的型付け      │
        │ → ネイティブ   │ │ チェーン解決   │ │ trick のメソッドを │
        │   Go コード生成│ │ breed A = B   │ │ learn で全て定義  │
        │ 型注釈なし     │ │ breed B = int │ │ すれば自動的に満足 │
        │ → meow.Value  │ │ → A は int    │ │                  │
        │   ボクシング   │ │   と互換      │ │                  │
        └───────────────┘ └──────────────┘ └──────────────────┘
```

## 優先度の高い拡張候補

Meow の関数型機能を強化する場合、以下の 3 つが最もインパクトが大きい。

### 1. 代数的データ型 (ADT) — Sum 型の導入

Sum 型がないため `Maybe` / `Either` / `Result` のような安全な値表現ができない。`kitty` を拡張して Sum 型を定義できるようにすると、エラーハンドリングやオプショナル値の表現力が大幅に向上する。

```meow
# 構想例
kitty Option {
  | Some(value)
  | None
}
```

### 2. ジェネリクス（パラメトリック多相）

型安全な汎用関数が書けず、全て `Value` 型に頼っている。ジェネリクスを導入すると `lick` / `picky` / `curl` などの高階関数や `trick` のメソッドシグネチャを型安全に記述できる。

```meow
# 構想例
meow identity(x T) T { bring x }

# ジェネリクス付き trick
trick Functor F {
    meow fmap(f paw(a) { b }, fa F) F
}
```

### 3. 高度なパターンマッチング

構造の分解やガード条件がないため、ADT を導入しても活かしきれない。`peek` を拡張して構造パターンやガードに対応すると表現力が飛躍的に向上する。

```meow
# 構想例
peek (option) {
  Some(x) sniff x > 0 => nya(x),
  None => nya("nothing"),
}
```

### 4. trick の拡張

現在の `trick` / `learn` は基本的な構造的型付けのみ。以下の拡張で Haskell の型クラスに近づく。

- **デフォルト実装**: trick 内でメソッドのデフォルト本体を定義可能にする
- **trick 継承**: `trick Ord extends Eq { ... }` のような trick 間の継承
- **プリミティブ型への learn**: `learn int { ... }` で組み込み型にもメソッドを追加
