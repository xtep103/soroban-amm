# Staking Contract

Located in [src/lib.rs](src/lib.rs).

An LP staking and rewards contract. Liquidity providers stake their LP tokens to
earn a separate reward token, distributed through a rewards-per-share accumulator
(the MasterChef pattern) so each claim is computed in O(1).

Stakers may optionally lock their stake for a fixed duration to receive a boost
multiplier on their reward share, modelled on Curve's veToken design. The boost
is applied to an *effective* staked amount used only for reward math; the actual
LP token balance is never altered. Boost scales linearly with the remaining lock
time between a configurable minimum (1x) and maximum, and locks can only be
extended, never shortened.

## How it fits into the protocol

The contract is standalone and composable. It holds two token addresses: an LP
token (staked in) and a reward token (paid out). The LP token can be the fungible
LP token minted by any V2 AMM pool, and the reward token can be any SEP-41 token,
so a protocol can incentivise liquidity in a specific pool without changes to the
AMM itself. The admin funds and distributes rewards; stakers stake, claim, and
unstake independently.

## Public functions

### Setup

| Function | Description |
|---|---|
| `initialize(lp_token, reward_token, admin)` | One-time setup with default boost and lock-duration bounds |
| `initialize_with_boost_config(lp_token, reward_token, admin, min_boost_scaled, max_boost_scaled, min_lock_duration_secs, max_lock_duration_secs)` | One-time setup with custom boost multipliers and lock-duration bounds |

### Staking and locking

| Function | Description |
|---|---|
| `stake(staker, amount)` | Stake LP tokens with no lock (1x boost) |
| `stake_locked(staker, amount, lock_duration_secs)` | Stake LP tokens with an optional lock duration for a boost multiplier |
| `lock(staker, amount, lock_duration_seconds)` | Escrow LP tokens for a fixed lock duration at a boosted reward rate |
| `extend_lock(staker, new_duration_seconds)` | Extend an existing lock forward in time only |
| `unlock(staker) → (amount, rewards)` | Withdraw the full staked amount and accrued rewards after the lock expires |
| `unstake(staker, amount) → (amount, rewards)` | Unstake LP tokens and claim pending rewards; panics if the lock has not expired |

### Rewards

| Function | Description |
|---|---|
| `add_rewards(admin, amount)` | Transfer reward tokens into the pool; admin only |
| `update_rewards(admin, new_rewards)` | Distribute new rewards across all stakers via the accumulator; admin only |
| `reconcile_reward_accounting(admin, offset, limit, done)` | Repair legacy reward accounting on a deployed pool (see issue #1045); admin only, paginated |
| `claim(staker) → rewards` | Claim accrued rewards without unstaking |
| `pending_rewards(staker) → i128` | Read a staker's unclaimed rewards |

### Views

| Function | Description |
|---|---|
| `get_pool_info() → PoolInfo` | Read pool state: token addresses, admin, total effective staked, undistributed reward balance, and accumulated rewards per share |
| `get_owed_rewards() → i128` | Read rewards already assigned to stakers but not yet paid out (#1045) |
| `get_staker_info(staker) → StakerInfo` | Read a staker's raw and effective amounts, rewards debt, lock expiry, and boost multiplier |
| `get_locked_position(staker) → LockedPosition` | Read a staker's locked amount, lock expiry, and boost multiplier |

### Admin

| Function | Description |
|---|---|
| `upgrade(admin, new_wasm_hash)` | Replace the contract WASM in place; `admin` must be the stored admin and sign. All staking state is preserved. See the [deployment runbook](../../docs/deployment-runbook.md#55-governance-concentrated-liquidity-oracle-aggregator-and-staking) |
