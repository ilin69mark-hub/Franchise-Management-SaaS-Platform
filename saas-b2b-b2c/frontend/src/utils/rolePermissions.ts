/**
 * Карта ролей → список ролей, которые может назначить пользователь.
 *
 * Если массив пустой – пользователь ничего не может назначать.
 */
export const assignableRolesMap: Record<string, string[]> = {
  // 1️⃣ Super‑admin – имеет право назначать любую роль
  super_admin: ['franchiser', 'franchiser_manager', 'dealer', 'salon_manager'],

  // 2️⃣ Франчайзи‑владелец (franchiser)
  //    – может ставить план менеджерам франчайзи, всем дилерам и их менеджерам салонов.
  //    На бек‑энде в canAssign разрешена роль `franchise_manager`
  //    (с «e»), поэтому здесь её и указываем.
  franchiser: ['franchise_manager', 'dealer', 'salon_manager'],

  // 3️⃣ Менеджер франчайзи (franchiser_manager)
  //    – может ставить план тем же, что и franchiser, но без возможности
  //    назначать самого franchiser.
  franchiser_manager: ['dealer', 'salon_manager'],

  // 4️⃣ Дилер (dealer) – может ставить только план менеджерам своих салонов.
  dealer: ['salon_manager'],

  // 5️⃣ Менеджер салона – ничего не может назначать.
  salon_manager: [],
};

/**
 * Возвращает массив ролей, которые текущий пользователь может назначать.
 * Если роль неизвестна – возвращаем пустой массив.
 */
export const getAssignableRoles = (role?: string | null): string[] => {
  if (!role) return [];
  return assignableRolesMap[role] ?? [];
};

/**
 * Проверка, может ли пользователь увидеть цель конкретного получателя.
 * (Используется в GoalList, если хотите фильтровать цели).
 */
export const canSeeGoal = (
  myRole: string | null,
  goalAssigneeRole: string,
  myId: string,
  goalAssigneeId: string,
) => {
  // 1️⃣ Своё собственное задание всегда видно
  if (goalAssigneeId === myId) return true;

  // 2️⃣ Если пользователь может назначать роль получателя → он её видит
  const assignable = getAssignableRoles(myRole);
  return assignable.includes(goalAssigneeRole);
};
