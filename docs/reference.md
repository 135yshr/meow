# Meow Language Reference

Meow 言語の全命令文・構文の一覧です。

## Keywords (キーワード)

| Meow | 意味 | 用例 |
|------|------|------|
| `nyan` | 変数宣言 | `nyan x = 42` |
| `meow` | 関数定義 | `meow add(a, b) { bring a + b }` |
| `bring` | 戻り値を返す | `bring x + 1` |
| `sniff` | 条件分岐 (if) | `sniff (x > 0) { ... }` |
| `scratch` | 条件分岐 (else) | `} scratch { ... }` |
| `purr` | ループ | `purr (i < 10) { ... }` |
| `paw` | ラムダ式 (無名関数) | `paw(x) { x * 2 }` |
| `nya` | 値を出力する | `nya("Hello!")` |
| `lick` | リストの各要素を変換する (map) | `lick(nums, paw(x) { x * 2 })` |
| `picky` | リストから条件に合う要素を選ぶ (filter) | `picky(nums, paw(x) { x > 0 })` |
| `curl` | リストを1つの値にまとめる (reduce) | `curl(nums, 0, paw(a, x) { a + x })` |
| `peek` | 値に応じて処理を分岐する (パターンマッチ) | `peek(v) { 0 => "zero", _ => "other" }` |
| `hiss` | エラーを発生させる | `hiss("something went wrong")` |
| `fetch` | インポート *(未実装)* | --- |
| `flaunt` | エクスポート *(未実装)* | --- |
| `yarn` | 真 (真偽値リテラル) | `nyan ok = yarn` |
| `hairball` | 偽 (真偽値リテラル) | `nyan ng = hairball` |
| `catnap` | 空値 (値が無いことを表す) | `nyan nothing = catnap` |

## Operators (演算子)

### 算術演算子

| 演算子 | 意味 | 用例 |
|--------|------|------|
| `+` | 加算 / 文字列結合 | `1 + 2`, `"a" + "b"` |
| `-` | 減算 / 単項マイナス | `5 - 3`, `-x` |
| `*` | 乗算 | `2 * 3` |
| `/` | 除算 | `10 / 2` |
| `%` | 剰余 | `10 % 3` |

### 比較演算子

| 演算子 | 意味 | 用例 |
|--------|------|------|
| `==` | 等価 | `x == 0` |
| `!=` | 非等価 | `x != 0` |
| `<` | 未満 | `x < 10` |
| `>` | 超過 | `x > 0` |
| `<=` | 以下 | `x <= 100` |
| `>=` | 以上 | `x >= 1` |

### 論理演算子

| 演算子 | 意味 | 用例 |
|--------|------|------|
| `&&` | 論理AND | `x > 0 && x < 10` |
| `\|\|` | 論理OR | `x == 0 \|\| x == 1` |
| `!` | 論理NOT | `!ok` |

### 特殊演算子

| 演算子 | 意味 | 用例 |
|--------|------|------|
| `\|=\|` | パイプ (土管) | `nums \|=\| lick(double)` |
| `..` | 範囲 | `1..10` |
| `=>` | マッチ腕 | `0 => "zero"` |
| `=` | 代入 | `nyan x = 1` |

## Literals (リテラル)

| 型 | 用例 | 説明 |
|----|------|------|
| 整数 | `42` | 10進整数 |
| 浮動小数点 | `3.14` | 小数点付き数値 |
| 文字列 | `"Hello, world!"` | ダブルクォート囲み、`\\` でエスケープ |

## Delimiters (区切り文字)

| 記号 | 意味 | 用例 |
|------|------|------|
| `(` `)` | 関数呼び出し / グループ化 | `add(1, 2)` |
| `{` `}` | ブロック | `meow f() { ... }` |
| `[` `]` | リスト / インデックス | `[1, 2, 3]`, `nums[0]` |
| `,` | 区切り | `add(a, b)` |

## Comments (コメント)

```
# 行コメント

-~ ブロックコメント
   複数行にまたがれます ~-
```

## Syntax Examples (構文例)

### 変数宣言

```
nyan x = 42
nyan greeting = "Hello!"
nyan pi = 3.14
nyan cats_are_great = yarn
nyan nothing = catnap
```

### 関数定義

```
meow add(a, b) {
  bring a + b
}

nya(add(1, 2))   # => 3
```

### 条件分岐

```
sniff (x > 0) {
  nya("positive")
} scratch sniff (x == 0) {
  nya("zero")
} scratch {
  nya("negative")
}
```

### ループ

```
nyan i = 0
purr (i < 10) {
  nya(i)
  i = i + 1
}
```

### ラムダ式

```
nyan double = paw(x) { x * 2 }
nya(double(5))   # => 10
```

### リスト操作

```
nyan nums = [1, 2, 3, 4, 5]

lick(nums, paw(x) { x * 2 })           # => [2, 4, 6, 8, 10]
picky(nums, paw(x) { x % 2 == 0 })     # => [2, 4]
curl(nums, 0, paw(acc, x) { acc + x })  # => 15
```

### パイプ

```
nyan double = paw(x) { x * 2 }
nums |=| lick(double)
```

### パターンマッチ

```
nyan result = peek(score) {
  0 => "zero",
  1..10 => "low",
  11..100 => "high",
  _ => "off the charts"
}
```
