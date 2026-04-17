from __future__ import annotations

from dataclasses import dataclass
from math import inf, isfinite
from typing import List

SENSE_LE = "<="
SENSE_GE = ">="
SENSE_EQ = "="

BIG_M = 1e6
MAX_ITER = 10_000


@dataclass
class Problem:
    c: List[float]
    A: List[List[float]]
    b: List[float]
    sense: List[str]


@dataclass
class IterationLP:
    k: int
    enter_var: int
    leave_var: int
    objective: float
    x: List[float]


@dataclass
class Result:
    x: List[float]
    objective: float
    iterations: int
    status: str
    trace: List[IterationLP]


def flip_sense(s: str) -> str:
    if s == SENSE_LE:
        return SENSE_GE
    if s == SENSE_GE:
        return SENSE_LE
    return s


def objective_value(c: List[float], x: List[float]) -> float:
    return sum(ci * xi for ci, xi in zip(c, x))


def extract_primal(tab: List[List[float]], basis: List[int], m: int, n: int, rhs_col: int) -> List[float]:
    x = [0.0] * n
    for i in range(m):
        if basis[i] < n:
            x[basis[i]] = tab[i][rhs_col]
    return x


def append_trace(
    trace: List[IterationLP],
    it: int,
    enter_col: int,
    leave_var: int,
    tab: List[List[float]],
    basis: List[int],
    m: int,
    n: int,
    rhs_col: int,
    c: List[float],
) -> None:
    x = extract_primal(tab, basis, m, n, rhs_col)
    trace.append(
        IterationLP(
            k=it,
            enter_var=(enter_col + 1) if enter_col >= 0 else 0,
            leave_var=(leave_var + 1) if leave_var >= 0 else 0,
            objective=objective_value(c, x),
            x=x,
        )
    )


def choose_entering(obj_row: List[float], last_col: int, eps: float) -> int:
    col = -1
    min_val = -eps
    for j in range(last_col):
        if obj_row[j] < min_val:
            min_val = obj_row[j]
            col = j
    return col


def choose_leaving(tab: List[List[float]], enter_col: int, m: int, rhs_col: int, eps: float) -> int:
    row = -1
    min_ratio = inf
    for i in range(m):
        a = tab[i][enter_col]
        if a <= eps:
            continue
        ratio = tab[i][rhs_col] / a
        if ratio < min_ratio - eps:
            min_ratio = ratio
            row = i
    return row


def pivot(tab: List[List[float]], pivot_row: int, pivot_col: int, rows: int, cols: int) -> None:
    pivot_val = tab[pivot_row][pivot_col]
    for j in range(cols):
        tab[pivot_row][j] /= pivot_val

    for i in range(rows):
        if i == pivot_row:
            continue
        factor = tab[i][pivot_col]
        if factor == 0.0:
            continue
        for j in range(cols):
            tab[i][j] -= factor * tab[pivot_row][j]


def solve_simplex(p: Problem, eps: float = 1e-6) -> Result:
    if eps <= 0 or not isfinite(eps):
        raise ValueError("eps must be positive finite number")

    m = len(p.b)
    n = len(p.c)
    if m == 0 or n == 0:
        raise ValueError("empty LP problem")
    if len(p.A) != m:
        raise ValueError("A rows mismatch")
    for i, row in enumerate(p.A):
        if len(row) != n:
            raise ValueError(f"A[{i}] cols mismatch")

    sense = p.sense[:] if p.sense else [SENSE_LE] * m
    if len(sense) != m:
        raise ValueError("sense len mismatch")

    A = [row[:] for row in p.A]
    b = p.b[:]

    for i in range(m):
        if b[i] < -eps:
            A[i] = [-v for v in A[i]]
            b[i] = -b[i]
            sense[i] = flip_sense(sense[i])
        if sense[i] not in (SENSE_LE, SENSE_GE, SENSE_EQ):
            raise ValueError(f"unsupported sense at row {i}: {sense[i]}")

    slack_count = sum(1 for s in sense if s == SENSE_LE)
    surplus_count = sum(1 for s in sense if s == SENSE_GE)
    art_count = sum(1 for s in sense if s in (SENSE_GE, SENSE_EQ))

    total_vars = n + slack_count + surplus_count + art_count
    rows = m + 1
    cols = total_vars + 1

    tab = [[0.0] * cols for _ in range(rows)]
    basis = [0] * m
    is_artificial = [False] * total_vars

    next_slack = n
    next_surplus = n + slack_count
    next_art = n + slack_count + surplus_count

    for i in range(m):
        tab[i][:n] = A[i][:]

        if sense[i] == SENSE_LE:
            tab[i][next_slack] = 1.0
            basis[i] = next_slack
            next_slack += 1
        elif sense[i] == SENSE_GE:
            tab[i][next_surplus] = -1.0
            tab[i][next_art] = 1.0
            basis[i] = next_art
            is_artificial[next_art] = True
            next_surplus += 1
            next_art += 1
        elif sense[i] == SENSE_EQ:
            tab[i][next_art] = 1.0
            basis[i] = next_art
            is_artificial[next_art] = True
            next_art += 1

        tab[i][cols - 1] = b[i]

    for j in range(n):
        tab[m][j] = -p.c[j]
    for j in range(total_vars):
        if is_artificial[j]:
            tab[m][j] = BIG_M

    for i in range(m):
        coef = tab[m][basis[i]]
        if abs(coef) <= eps:
            continue
        for j in range(cols):
            tab[m][j] -= coef * tab[i][j]

    trace: List[IterationLP] = []
    it = 0
    append_trace(trace, it, -1, -1, tab, basis, m, n, cols - 1, p.c)

    while it < MAX_ITER:
        enter_col = choose_entering(tab[m], cols - 1, eps)
        if enter_col == -1:
            infeasible = False
            for i in range(m):
                if is_artificial[basis[i]] and tab[i][cols - 1] > eps:
                    infeasible = True
                    break

            x = [0.0] * n
            for i in range(m):
                if basis[i] < n:
                    x[basis[i]] = tab[i][cols - 1]

            return Result(
                x=x,
                objective=objective_value(p.c, x),
                iterations=it,
                status="infeasible" if infeasible else "optimal",
                trace=trace,
            )

        leave_row = choose_leaving(tab, enter_col, m, cols - 1, eps)
        if leave_row == -1:
            return Result(
                x=[],
                objective=inf,
                iterations=it,
                status="unbounded",
                trace=trace,
            )

        leave_var = basis[leave_row]
        pivot(tab, leave_row, enter_col, rows, cols)
        basis[leave_row] = enter_col
        it += 1
        append_trace(trace, it, enter_col, leave_var, tab, basis, m, n, cols - 1, p.c)

    raise RuntimeError("simplex reached iteration limit")


def build_variant_problem() -> Problem:
    # F = 7x1 - 2x2 -> max
    # 5x1 - 2x2 <= 3
    # x1 + x2 >= 1
    # -3x1 + x2 <= 3
    # 2x1 + x2 <= 4
    # x1, x2 >= 0
    return Problem(
        c=[7.0, -2.0],
        A=[
            [5.0, -2.0],
            [1.0, 1.0],
            [-3.0, 1.0],
            [2.0, 1.0],
        ],
        b=[3.0, 1.0, 3.0, 4.0],
        sense=[SENSE_LE, SENSE_GE, SENSE_LE, SENSE_LE],
    )


def main() -> None:
    prob = build_variant_problem()
    res = solve_simplex(prob, eps=1e-6)

    print("Simplex (Python) for the provided variant")
    print(f"status     = {res.status}")
    if len(res.x) >= 2:
        print(f"x*         = ({res.x[0]:.6f}; {res.x[1]:.6f})")
    print(f"F(x*)      = {res.objective:.6f}")
    print(f"iterations = {res.iterations}")


if __name__ == "__main__":
    main()
