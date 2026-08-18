import 'package:flutter_test/flutter_test.dart';

import 'package:aslibhaav_app/main.dart';

void main() {
  testWidgets('computes the effective cost from the default scheme', (tester) async {
    await tester.pumpWidget(const AslibhaavApp());
    await tester.tap(find.text('Show the real price'));
    await tester.pump();
    // Default scheme (100, 10%, 5%, buy 10 get 2, freight 120 extra) => ₹81.25
    expect(find.text('₹81.25'), findsOneWidget);
  });

  test('effectiveCost matches the Go oracle for the full scheme', () {
    final b = effectiveCost(
      basePrice: 100, tradePct: 10, cashPct: 5,
      buyQty: 10, getQty: 2, freight: 120, freightExtra: true,
    );
    expect(b.effective, closeTo(81.25, 1e-9));
  });
}
